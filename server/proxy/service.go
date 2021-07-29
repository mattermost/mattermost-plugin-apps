// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upkubeless"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	mm        *pluginapi.Client
	log       utils.Logger
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

	AddLocalManifest(actingUserID string, m *apps.Manifest) (md.MD, error)
	AppIsEnabled(app *apps.App) bool
	EnableApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (md.MD, error)
	DisableApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (md.MD, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []*apps.App
	GetListedApps(filter string, includePluginApps bool) []*apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error)
	InstallApp(_ apps.AppID, _ mmclient.Client, sessionID string, _ *apps.Context, trusted bool, secret string, _ apps.DeployType) (*apps.App, md.MD, error)
	SynchronizeInstalledApps() error
	UninstallApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (md.MD, error)

	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
}

var _ Service = (*Proxy)(nil)

func NewService(mm *pluginapi.Client, log utils.Logger, conf config.Service, store *store.Service, mutex *cluster.Mutex, httpOut httpout.Service) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		mm:               mm,
		log:              log,
		conf:             conf,
		store:            store,
		callOnceMutex:    mutex,
		httpOut:          httpOut,
	}
}

func (p *Proxy) Configure(conf config.Config) error {
	newUpstream := func(dtype apps.DeployType, makeUpstream func() (upstream.Upstream, error)) {
		if isDeploySupported(conf, dtype) == nil {
			up, err := makeUpstream()
			if err != nil {
				p.mm.Log.Debug("failed to initialize upstream", "error", err.Error(), "app_type", dtype)
			} else {
				p.upstreams.Store(dtype, up)
			}
		} else {
			p.upstreams.Delete(dtype)
		}
	}

	newUpstream(apps.DeployHTTP, func() (upstream.Upstream, error) {
		return uphttp.NewUpstream(p.httpOut), nil
	})
	newUpstream(apps.DeployAWSLambda, func() (upstream.Upstream, error) {
		return upaws.MakeUpstream(conf.AWSAccessKey, conf.AWSSecretKey, conf.AWSRegion, conf.AWSS3Bucket, p.log)
	})
	newUpstream(apps.DeployPlugin, func() (upstream.Upstream, error) {
		return upplugin.NewUpstream(&p.mm.Plugin), nil
	})
	newUpstream(apps.DeployKubeless, func() (upstream.Upstream, error) {
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
