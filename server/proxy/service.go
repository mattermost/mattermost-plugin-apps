// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost-plugin-api/cluster"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
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
	config.Configurable

	Call(sessionID, actingUserID string, creq *apps.CallRequest) *apps.ProxyCallResponse
	CompleteRemoteOAuth2(sessionID, actingUserID string, appID apps.AppID, urlValues map[string]interface{}) error
	GetStatic(appID apps.AppID, path string) (io.ReadCloser, int, error)
	GetBindings(sessionID, actingUserID string, cc *apps.Context) ([]*apps.Binding, error)
	GetRemoteOAuth2ConnectURL(sessionID, actingUserID string, appID apps.AppID) (string, error)
	Notify(cc *apps.Context, subj apps.Subject) error
	NotifyRemoteWebhook(app *apps.App, data []byte, path string) error

	AddLocalManifest(actingUserID string, m *apps.Manifest) (string, error)
	AppIsEnabled(app *apps.App) bool
	EnableApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (string, error)
	DisableApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (string, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []*apps.App
	GetListedApps(filter string, includePluginApps bool) []*apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
	InstallApp(client mmclient.Client, sessionID string, cc *apps.Context, trusted bool, secret string) (*apps.App, string, error)
	SynchronizeInstalledApps() error
	UninstallApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (string, error)

	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
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
	newUpstream := func(appType apps.AppType, makeUpstream func() (upstream.Upstream, error)) {
		if isAppTypeSupported(conf, appType) == nil {
			up, err := makeUpstream()
			if err != nil {
				p.conf.Logger().Debugw("failed to initialize upstream", "error", err.Error(), "app_type", appType)
			} else {
				p.upstreams.Store(appType, up)
			}
		} else {
			p.upstreams.Delete(appType)
		}
	}

	newUpstream(apps.AppTypeHTTP, func() (upstream.Upstream, error) {
		return uphttp.NewUpstream(p.httpOut), nil
	})
	newUpstream(apps.AppTypeAWSLambda, func() (upstream.Upstream, error) {
		return upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, log)
	})
	newUpstream(apps.AppTypePlugin, func() (upstream.Upstream, error) {
		return upplugin.NewUpstream(&mm.Plugin), nil
	})
	newUpstream(apps.AppTypeKubeless, func() (upstream.Upstream, error) {
		return upkubeless.MakeUpstream()
	})
	return nil
}

func (p *Proxy) AddBuiltinUpstream(appID apps.AppID, up upstream.Upstream) {
	if p.builtinUpstreams == nil {
		p.builtinUpstreams = map[apps.AppID]upstream.Upstream{}
	}
	p.builtinUpstreams[appID] = up
}

func WriteCallError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		Type:      apps.CallResponseTypeError,
		ErrorText: err.Error(),
	})
}
