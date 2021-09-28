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
	"github.com/mattermost/mattermost-server/v6/model"

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

// Admin defines the REST API methods to manipulate Apps.
type Admin interface {
	DisableApp(Incoming, apps.Context, apps.AppID) (string, error)
	EnableApp(Incoming, apps.Context, apps.AppID) (string, error)
	InstallApp(_ Incoming, _ apps.Context, _ apps.AppID, trustedApp bool, secret string) (*apps.App, string, error)
	UninstallApp(Incoming, apps.Context, apps.AppID) (string, error)
}

// Invoker implements operations that invoke the Apps.
type Invoker interface {
	// REST API methods used by user agents (mobile, desktop, web).
	Call(Incoming, apps.CallRequest) apps.ProxyCallResponse
	CompleteRemoteOAuth2(_ Incoming, _ apps.AppID, urlValues map[string]interface{}) error
	GetBindings(Incoming, apps.Context) ([]apps.Binding, error)
	GetRemoteOAuth2ConnectURL(Incoming, apps.AppID) (string, error)
	GetStatic(_ apps.AppID, path string) (io.ReadCloser, int, error)
}

// Notifier implements user-less notification sinks.
type Notifier interface {
	Notify(apps.Context, apps.Subject) error
	NotifyRemoteWebhook(app apps.App, data []byte, path string) error
	NotifyMessageHasBeenPosted(*model.Post, apps.Context) error
	NotifyUserHasJoinedChannel(apps.Context) error
	NotifyUserHasLeftChannel(apps.Context) error
	NotifyUserHasJoinedTeam(apps.Context) error
	NotifyUserHasLeftTeam(apps.Context) error
}

// Internal implements go API used by other packages.
type Internal interface {
	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
	AddLocalManifest(m apps.Manifest) (string, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []apps.App
	GetListedApps(filter string, includePluginApps bool) []apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
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

	p.initUpstream(apps.AppTypeHTTP, conf, log, func() (upstream.Upstream, error) {
		return uphttp.NewUpstream(p.httpOut, conf.DeveloperMode), nil
	})
	p.initUpstream(apps.AppTypeAWSLambda, conf, log, func() (upstream.Upstream, error) {
		return upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
	})
	p.initUpstream(apps.AppTypePlugin, conf, log, func() (upstream.Upstream, error) {
		return upplugin.NewUpstream(&mm.Plugin), nil
	})
	p.initUpstream(apps.AppTypeKubeless, conf, log, func() (upstream.Upstream, error) {
		return upkubeless.MakeUpstream()
	})
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

func (p *Proxy) initUpstream(typ apps.AppType, newConfig config.Config, log utils.Logger, makef func() (upstream.Upstream, error)) {
	if err := isAppTypeSupported(newConfig, typ); err == nil {
		var up upstream.Upstream
		up, err = makef()
		switch {
		case errors.Cause(err) == utils.ErrNotFound:
			log.WithError(err).Debugf("Skipped %s upstream: not configured.", typ)
		case err != nil:
			log.WithError(err).Errorf("Failed to initialize %s upstream.", typ)
		default:
			p.upstreams.Store(typ, up)
			log.Debugf("Initialized %s upstream.", typ)
		}
	} else {
		p.upstreams.Delete(typ)
		log.Debugf("Upstream %s is not supported, cause: %v", typ, err)
	}
}
