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
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
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
	aws           awsclient.Client
	s3AssetBucket string
}

type Service interface {
	Call(apps.SessionToken, *apps.CallRequest) *apps.CallResponse
	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)
	GetBindings(apps.SessionToken, *apps.Context) ([]*apps.Binding, error)
	Notify(cc *apps.Context, subj apps.Subject) error

	AddLocalManifest(*apps.Context, apps.SessionToken, *apps.Manifest) (md.MD, error)
	GetInstalledApp(appID apps.AppID) (*apps.App, error)
	GetInstalledApps() []*apps.App
	GetListedApps(filter string) []*apps.ListedApp
	GetManifest(appID apps.AppID) (*apps.Manifest, error)
	InstallApp(*apps.Context, apps.SessionToken, *apps.InInstallApp) (*apps.App, md.MD, error)
	SynchronizeInstalledApps() error
	UninstallApp(appID apps.AppID) error

	AddBuiltinUpstream(apps.AppID, upstream.Upstream)
}

var _ Service = (*Proxy)(nil)

func NewService(mm *pluginapi.Client, aws awsclient.Client, conf config.Service, store *store.Service, s3AssetBucket string, mutex *cluster.Mutex) *Proxy {
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
