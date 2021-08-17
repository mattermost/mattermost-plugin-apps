// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	conf      config.Service
	store     *store.Service
	httpOut   httpout.Service
	upstreams sync.Map // key: apps.AppID, value upstream.Upstream
}

type Service interface {
	// To update on configuration changes
	config.Configurable

	// Admin REST API methods.
	DisableApp(Incoming, apps.Context, apps.AppID) (string, error)
	EnableApp(Incoming, apps.Context, apps.AppID) (string, error)
	InstallApp(_ Incoming, _ apps.Context, _ apps.AppID, _ apps.DeployType, trusted bool, secret string) (*apps.App, string, error)
	UninstallApp(Incoming, apps.Context, apps.AppID) (string, error)

	// REST API methods used by user agents (mobile, desktop, web).
	Call(Incoming, apps.CallRequest) apps.ProxyCallResponse
	CompleteRemoteOAuth2(_ Incoming, _ apps.AppID, urlValues map[string]interface{}) error
	GetBindings(Incoming, apps.Context) ([]apps.Binding, error)
	GetRemoteOAuth2ConnectURL(Incoming, apps.AppID) (string, error)
	GetStatic(_ apps.AppID, path string) (io.ReadCloser, int, error)

	// User-less notification sinks.
	Notify(apps.Context, apps.Subject) error
	NotifyRemoteWebhook(app apps.App, data []byte, path string) error
	NotifyMessageHasBeenPosted(*model.Post, apps.Context) error
	NotifyUserHasJoinedChannel(apps.Context) error
	NotifyUserHasLeftChannel(apps.Context) error
	NotifyUserHasJoinedTeam(apps.Context) error
	NotifyUserHasLeftTeam(apps.Context) error

	// Internal go API used by other packages.
	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
	AddLocalManifest(m apps.Manifest) (string, error)
	CanDeploy(deployType apps.DeployType) (allowed, usable bool)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []apps.App
	GetListedApps(filter string, includePluginApps bool) []apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
	SynchronizeInstalledApps() error
}

var _ Service = (*Proxy)(nil)

func NewService(conf config.Service, store *store.Service, mutex *cluster.Mutex, httpOut httpout.Service) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		conf:             conf,
		store:            store,
		callOnceMutex:    mutex,
		httpOut:          httpOut,
	}
}

func (p *Proxy) Configure(conf config.Config) error {
	_, mm, log := p.conf.Basic()

	if allowed, _ := p.CanDeploy(apps.DeployHTTP); allowed {
		p.upstreams.Store(apps.DeployHTTP, uphttp.NewUpstream(p.httpOut))
	} else {
		p.upstreams.Delete(apps.DeployHTTP)
	}

	if allowed, _ := p.CanDeploy(apps.DeployAWSLambda); allowed {
		up, err := upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
		if err != nil {
			return errors.Wrap(err, "failed to initialize AWS upstream")
		}
		p.upstreams.Store(apps.DeployAWSLambda, up)
	} else {
		p.upstreams.Delete(apps.DeployAWSLambda)
	}

	if allowed, _ := p.CanDeploy(apps.DeployPlugin); allowed {
		p.upstreams.Store(apps.DeployPlugin, upplugin.NewUpstream(&mm.Plugin))
	} else {
		p.upstreams.Delete(apps.DeployPlugin)
	}

	if allowed, _ := p.CanDeploy(apps.DeployKubeless); allowed {
		up, err := upkubeless.MakeUpstream()
		if err != nil {
			return errors.Wrap(err, "failed to initialize Kubeless upstream")
		}
		p.upstreams.Store(apps.DeployKubeless, up)
	} else {
		p.upstreams.Delete(apps.DeployKubeless)
	}

	return nil
}

// CanDeploy returns the availability of deployType.  allowed indicates that the
// type can be used in the current configuration. usable indicates that it is
// configured and can be accessed, or deployed to.
func (p *Proxy) CanDeploy(deployType apps.DeployType) (allowed, usable bool) {
	_, usable = p.upstreams.Load(deployType)

	conf := p.conf.Get()
	supportedTypes := []apps.DeployType{}

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
			apps.DeployKubeless)

	default:
		return false, false
	}

	for _, t := range supportedTypes {
		if deployType == t {
			return true, usable
		}
	}
	return false, false
}

func CanDeploy(p Service, deployType apps.DeployType) error {
	_, canDeploy := p.CanDeploy(deployType)
	if !canDeploy {
		return errors.Errorf("%s app deployment is not configured on this instance of Mattermost", deployType)
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

func WriteCallError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		Type:      apps.CallResponseTypeError,
		ErrorText: err.Error(),
	})
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
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
