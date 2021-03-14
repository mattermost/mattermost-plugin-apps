// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type BuiltinApp interface {
	// may also implement upstream.Upstream
	App() *apps.App
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

	AddBuiltinUpstream(apps.AppID, Upstream)
}

type proxy struct {
	// Manifests contains all relevant manifests. For V1, the entire list is
	// cached in memory, and loaded on startup.
	Manifests map[apps.AppID]*apps.Manifest

	// Built-in Apps are linked in Go and invoked directly. The list is
	// initialized on startup, and need not be synchronized. Built-in apps do
	// not need manifests.
	builtinUpstreams map[apps.AppID]upstream.Upstream

	mm    *pluginapi.Client
	conf  config.Service
	store *store.Service
	aws   aws.Service
}

var _ Service = (*proxy)(nil)

func NewService(mm *pluginapi.Client, aws aws.Service, conf config.Service, store *store.Service) Service {
	return &proxy{
		mm:    mm,
		conf:  conf,
		store: store,
		aws:   aws,
	}
}

func (p *proxy) InitBuiltinApps(builtinApps ...BuiltinApp) {
	for _, b := range builtinApps {
		app := b.App()
		if app != nil {
			p.store.App.InitBuiltin(app)
		}

		up, ok := b.(upstream.Upstream)
		if ok {
			if p.builtinUpstreams == nil {
				p.builtinUpstreams = map[apps.AppID]upstream.Upstream{}
			}
			p.builtinUpstreams[app.AppID] = up
		}
	}
}
