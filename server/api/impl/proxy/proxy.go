// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"errors"
	"net/http"

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
	if app.Manifest.RootURL == "" {
		return nil, errors.New("only built-in and remote http upstreams are supported, hosted AWS Lambda coming soon")
	}

	// TODO: support AWS Lambda upstream
	up = uphttp.NewUpstream(app)

	return up, nil
}

func (p *Proxy) ProvisionBuiltIn(appID api.AppID, up api.Upstream) {
	if p.builtIn == nil {
		p.builtIn = map[api.AppID]api.Upstream{}
	}
	p.builtIn[appID] = up
}

func WriteCallError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(api.CallResponse{
		Type:      api.CallResponseTypeError,
		ErrorText: err.Error(),
	})
}
