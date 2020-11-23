// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream/uphttp"
)

type Proxy struct {
	builtIn map[api.AppID]api.Upstream

	mm    *pluginapi.Client
	conf  api.Configurator
	store api.Store
}

var _ api.Proxy = (*Proxy)(nil)

func NewProxy(mm *pluginapi.Client, conf api.Configurator, store api.Store) *Proxy {
	return &Proxy{nil, mm, conf, store}
}

func (p *Proxy) upstreamForApp(app *api.App) (api.Upstream, error) {
	var up api.Upstream

	if len(p.builtIn) > 0 {
		up = p.builtIn[app.Manifest.AppID]
		if up != nil {
			return up, nil
		}
	}
	if app.Manifest.RemoteRootURL == "" {
		return nil, errors.New("only built-in and remote http upstreams are supported, hosted AWS Lambda coming soon")
	}

	// TODO: support AWS Lambda upstream
	up = uphttp.NewHTTPUpstream(app.Manifest.AppID, app.Manifest.RemoteRootURL, app.Secret)

	return up, nil
}

func (p *Proxy) DebugBuiltInApp(appID api.AppID, up api.Upstream) {
	if p.builtIn == nil {
		p.builtIn = map[api.AppID]api.Upstream{}
	}
	p.builtIn[appID] = up
}
