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
	"github.com/mattermost/mattermost-plugin-apps/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Proxy struct {
	callOnceMutex *cluster.Mutex

	builtinUpstreams map[apps.AppID]upstream.Upstream

	mm            *pluginapi.Client
	conf          config.Service
	store         *store.Service
	aws           aws.Client
	s3AssetBucket string
}

type Service interface {
	Call(apps.SessionToken, *apps.CallRequest) *apps.CallResponse
	CompleteRemoteOAuth2(appID apps.AppID, actingUserID, token string, urlValues map[string]interface{}) error
	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)
	GetBindings(apps.SessionToken, *apps.Context) ([]*apps.Binding, error)
	GetRemoteOAuth2RedirectURL(appID apps.AppID, actingUserID, token string) (string, error)
	Notify(cc *apps.Context, subj apps.Subject) error

	AddLocalManifest(actingUserID string, _ apps.SessionToken, _ *apps.Manifest) (md.MD, error)
	AppIsEnabled(app *apps.App) bool
	DisableApp(cc *apps.Context, app *apps.App) (md.MD, error)
	EnableApp(cc *apps.Context, app *apps.App) (md.MD, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []*apps.App
	GetListedApps(filter string) []*apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	SynchronizeInstalledApps() error
	UninstallApp(appID apps.AppID, sessionToken apps.SessionToken, actingUserID string) error

	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
}

var _ Service = (*Proxy)(nil)

func NewService(mm *pluginapi.Client, aws aws.Client, conf config.Service, store *store.Service, s3AssetBucket string, mutex *cluster.Mutex) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]upstream.Upstream{},
		mm:               mm,
		conf:             conf,
		store:            store,
		aws:              aws,
		s3AssetBucket:    s3AssetBucket,
		callOnceMutex:    mutex,
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
