// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream/upawslambda"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream/uphttp"
)

type Proxy struct {
	builtIn map[api.AppID]api.Upstream

	mm        *pluginapi.Client
	conf      api.Configurator
	store     api.Store
	awsClient *aws.Client
}

var _ api.Proxy = (*Proxy)(nil)

func NewProxy(mm *pluginapi.Client, awsClient *aws.Client, conf api.Configurator, store api.Store) *Proxy {
	return &Proxy{nil, mm, conf, store, awsClient}
}

func (p *Proxy) Call(debugSessionToken api.SessionToken, c *api.Call) *api.CallResponse {
	app, err := p.store.LoadApp(c.Context.AppID)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}

	expander := p.newExpander(c.Context, p.mm, p.conf, p.store, debugSessionToken)
	cc, err := expander.ExpandForApp(app, c.Expand)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	clone := *c
	clone.Context = cc

	return up.Call(&clone)
}

func (p *Proxy) Notify(cc *api.Context, subj api.Subject) error {
	subs, err := p.store.LoadSubs(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := p.newExpander(cc, p.mm, p.conf, p.store, "")

	notify := func(sub *api.Subscription) error {
		call := sub.Call
		if call == nil {
			return errors.New("nothing to call")
		}
		app, err := p.store.LoadApp(sub.AppID)
		if err != nil {
			return err
		}
		call.Context, err = expander.ExpandForApp(app, call.Expand)
		if err != nil {
			return err
		}
		call.Context.Subject = subj

		up, err := p.upstreamForApp(app)
		if err != nil {
			return err
		}
		go func() {
			_ = up.Notify(call)
		}()
		return nil
	}

	for _, sub := range subs {
		err := notify(sub)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Proxy) upstreamForApp(app *api.App) (api.Upstream, error) {
	switch app.Manifest.Type {
	case api.AppTypeHTTP:
		return uphttp.NewUpstream(app), nil

	case api.AppTypeAWSLambda:
		return upawslambda.NewUpstream(app, p.awsClient), nil

	case api.AppTypeBuiltin:
		if len(p.builtIn) == 0 {
			return nil, errors.Errorf("builtin app not found: %s", app.Manifest.AppID)
		}
		up := p.builtIn[app.Manifest.AppID]
		if up == nil {
			return nil, errors.Errorf("builtin app not found: %s", app.Manifest.AppID)
		}
		return up, nil

	default:
		return nil, errors.Errorf("not a valid app type: %s", app.Manifest.Type)
	}
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
