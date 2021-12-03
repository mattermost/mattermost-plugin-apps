// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upopenfaas"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	conf           config.Service
	store          *store.Service
	httpOut        httpout.Service
	upstreams      sync.Map // key: apps.AppID, value upstream.Upstream
	sessionService session.Service
	appservices    appservices.Service

	// expandClientOverride is set by the tests to use the mock client
	expandClientOverride mmclient.Client
}

// Admin defines the REST API methods to manipulate Apps.
type Admin interface {
	DisableApp(*incoming.Request, apps.Context, apps.AppID) (string, error)
	EnableApp(*incoming.Request, apps.Context, apps.AppID) (string, error)
	InstallApp(_ *incoming.Request, _ apps.Context, _ apps.AppID, _ apps.DeployType, trustedApp bool, secret string) (*apps.App, string, error)
	UpdateAppListing(*incoming.Request, appclient.UpdateAppListingRequest) (*apps.Manifest, error)
	UninstallApp(*incoming.Request, apps.Context, apps.AppID) (string, error)
}

// Invoker implements operations that invoke the Apps.
type Invoker interface {
	// REST API methods used by user agents (mobile, desktop, web).
	Call(*incoming.Request, apps.CallRequest) CallResponse
	CompleteRemoteOAuth2(_ *incoming.Request, _ apps.AppID, urlValues map[string]interface{}) error
	GetBindings(*incoming.Request, apps.Context) ([]apps.Binding, error)
	GetRemoteOAuth2ConnectURL(*incoming.Request, apps.AppID) (string, error)
	GetStatic(_ *incoming.Request, _ apps.AppID, path string) (io.ReadCloser, int, error)
}

// Notifier implements user-less notification sinks.
type Notifier interface {
	Notify(apps.Context, apps.Subject) error
	NotifyRemoteWebhook(*incoming.Request, apps.AppID, apps.HTTPCallRequest) error
	NotifyMessageHasBeenPosted(*model.Post, apps.Context) error
	NotifyUserHasJoinedChannel(apps.Context) error
	NotifyUserHasLeftChannel(apps.Context) error
	NotifyUserHasJoinedTeam(apps.Context) error
	NotifyUserHasLeftTeam(apps.Context) error
}

// Internal implements go API used by other packages.
type Internal interface {
	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
	CanDeploy(deployType apps.DeployType) (allowed, usable bool)
	GetAppBindings(r *incoming.Request, cc apps.Context, app apps.App) []apps.Binding
	GetInstalledApp(r *incoming.Request, appID apps.AppID) (*apps.App, error)
	GetInstalledApps(r *incoming.Request) []apps.App
	GetListedApps(r *incoming.Request, filter string, includePluginApps bool) []apps.ListedApp
	GetManifest(r *incoming.Request, appID apps.AppID) (*apps.Manifest, error)
	SynchronizeInstalledApps() error
}

type Service interface {
	// To update on configuration changes
	config.Configurable

	Admin
	Internal
	Invoker
	Notifier
}

var _ Service = (*Proxy)(nil)

func NewService(conf config.Service, store *store.Service, mutex *cluster.Mutex, httpOut httpout.Service, session session.Service, appservices appservices.Service) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		conf:             conf,
		store:            store,
		callOnceMutex:    mutex,
		httpOut:          httpOut,
		sessionService:   session,
		appservices:      appservices,
	}
}

func (p *Proxy) Configure(conf config.Config, log utils.Logger) error {
	mm := p.conf.MattermostAPI()

	p.initUpstream(apps.DeployHTTP, conf, log, func() (upstream.Upstream, error) {
		return uphttp.NewUpstream(p.httpOut, conf.DeveloperMode, uphttp.AppRootURL), nil
	})
	p.initUpstream(apps.DeployAWSLambda, conf, log, func() (upstream.Upstream, error) {
		return upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
	})
	p.initUpstream(apps.DeployPlugin, conf, log, func() (upstream.Upstream, error) {
		return upplugin.NewUpstream(&mm.Plugin), nil
	})
	p.initUpstream(apps.DeployKubeless, conf, log, func() (upstream.Upstream, error) {
		// Kubeless is not longer supported: https://mattermost.atlassian.net/browse/MM-40011
		// return upkubeless.MakeUpstream()
		return nil, nil
	})
	p.initUpstream(apps.DeployOpenFAAS, conf, log, func() (upstream.Upstream, error) {
		return upopenfaas.MakeUpstream(p.httpOut, conf.DeveloperMode)
	})
	return nil
}

// CanDeploy returns the availability of deployType. allowed indicates that the
// type can be used in the current configuration. usable indicates that it is
// configured and can be accessed, or deployed to.
func (p *Proxy) CanDeploy(deployType apps.DeployType) (allowed, usable bool) {
	return p.canDeploy(p.conf.Get(), deployType)
}

func (p *Proxy) canDeploy(conf config.Config, deployType apps.DeployType) (allowed, usable bool) {
	_, usable = p.upstreams.Load(deployType)

	supportedTypes := apps.DeployTypes{}

	// Initialize with the set supported in all configurations.
	supportedTypes = append(supportedTypes,
		apps.DeployAWSLambda,
		apps.DeployBuiltin,
		apps.DeployPlugin,
	)

	switch {
	case conf.DeveloperMode:
		// In dev mode support any deploy type.
		return true, usable

	case conf.MattermostCloudMode:
		// Nothing else in Mattermost Cloud mode.

	case !conf.MattermostCloudMode:
		// Add more deploy types in self-managed mode.
		supportedTypes = append(supportedTypes,
			apps.DeployHTTP,
			apps.DeployKubeless,
			apps.DeployOpenFAAS,
		)
	}

	for _, t := range supportedTypes {
		if deployType == t {
			return true, usable
		}
	}
	return false, false
}

// CanDeploy returns the availability of deployType. allowed indicates that the
// type can be used in the current configuration. usable indicates that it is
// configured and can be accessed, or deployed to.
func CanDeploy(p Service, deployType apps.DeployType) error {
	_, canDeploy := p.CanDeploy(deployType)
	if !canDeploy {
		return errors.Errorf("deployment type %q is not configured on this Mattermost server", deployType)
	}
	return nil
}

func (p *Proxy) AddBuiltinUpstream(appID apps.AppID, up upstream.Upstream) {
	if p.builtinUpstreams == nil {
		p.builtinUpstreams = map[apps.AppID]upstream.Upstream{}
	}
	p.builtinUpstreams[appID] = up
	p.store.App.InitBuiltin()
}

func (p *Proxy) upstreamForApp(app apps.App) (upstream.Upstream, error) {
	if app.DeployType == apps.DeployBuiltin {
		u, ok := p.builtinUpstreams[app.AppID]
		if !ok {
			return nil, errors.Wrapf(utils.ErrNotFound, "no builtin %s", app.AppID)
		}
		return u, nil
	}

	err := CanDeploy(p, app.DeployType)
	if err != nil {
		return nil, err
	}

	upv, ok := p.upstreams.Load(app.DeployType)
	if !ok {
		return nil, utils.NewInvalidError("invalid or unsupported upstream type: %s", app.DeployType)
	}
	up, ok := upv.(upstream.Upstream)
	if !ok {
		return nil, utils.NewInvalidError("invalid Upstream for: %s", app.DeployType)
	}
	return up, nil
}

func (p *Proxy) initUpstream(typ apps.DeployType, newConfig config.Config, log utils.Logger, makef func() (upstream.Upstream, error)) {
	if allowed, _ := p.canDeploy(newConfig, typ); allowed {
		up, err := makef()
		switch {
		case errors.Cause(err) == utils.ErrNotFound:
			log.WithError(err).Debugf("Skipped %q upstream: not configured.", typ)
		case err != nil:
			log.WithError(err).Errorf("Failed to initialize %q upstream.", typ)
		default:
			p.upstreams.Store(typ, up)
			log.Debugf("Initialized %q upstream.", typ)
		}
	} else {
		p.upstreams.Delete(typ)
		log.Debugf("Deploy type %q is not configured on this Mattermost server", typ)
	}
}
