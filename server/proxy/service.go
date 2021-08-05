// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpout"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	mm            *pluginapi.Client
	log           utils.Logger
	conf          config.Service
	store         *store.Service
	aws           upaws.Client
	httpOut       httpout.Service
	s3AssetBucket string
}

type Service interface {
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
	InstallApp(client mmclient.Client, sessionID string, cc *apps.Context, trusted bool, secret, pluginID string) (*apps.App, string, error)
	SynchronizeInstalledApps() error
	UninstallApp(client mmclient.Client, sessionID string, cc *apps.Context, appID apps.AppID) (string, error)

	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
}

var _ Service = (*Proxy)(nil)

func NewService(mm *pluginapi.Client, log utils.Logger, conf config.Service, aws upaws.Client, s3AssetBucket string, store *store.Service, mutex *cluster.Mutex, httpOut httpout.Service) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		mm:               mm,
		log:              log,
		conf:             conf,
		store:            store,
		aws:              aws,
		s3AssetBucket:    s3AssetBucket,
		callOnceMutex:    mutex,
		httpOut:          httpOut,
	}
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
