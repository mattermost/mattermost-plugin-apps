// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"net/http"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream/upawslambda"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream/uphttp"
)

type Proxy struct {
	builtinUpstreams map[apps.AppID]api.Upstream

	mm            *pluginapi.Client
	conf          api.Configurator
	store         api.Store
	aws           awsclient.Client
	s3AssetBucket string
}

var _ api.Proxy = (*Proxy)(nil)

func NewProxy(mm *pluginapi.Client, aws awsclient.Client, conf api.Configurator, store api.Store, s3AssetBucket string) *Proxy {
	return &Proxy{
		builtinUpstreams: map[apps.AppID]api.Upstream{},
		mm:               mm,
		conf:             conf,
		store:            store,
		aws:              aws,
		s3AssetBucket:    s3AssetBucket,
	}
}

func (p *Proxy) Call(debugSessionToken apps.SessionToken, c *apps.CallRequest) *apps.CallResponse {
	app, err := p.store.App().Get(c.Context.AppID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	expander := p.newExpander(c.Context, p.mm, p.conf, p.store, debugSessionToken)
	cc, err := expander.ExpandForApp(app, c.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	clone := *c
	clone.Context = cc

	return upstream.Call(up, &clone)
}

func (p *Proxy) Notify(cc *apps.Context, subj apps.Subject) error {
	subs, err := p.store.Sub().Get(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := p.newExpander(cc, p.mm, p.conf, p.store, "")

	notify := func(sub *apps.Subscription) error {
		call := sub.Call
		if call == nil {
			return errors.New("nothing to call")
		}

		callRequest := &apps.CallRequest{Call: *call}
		app, err := p.store.App().Get(sub.AppID)
		if err != nil {
			return err
		}
		callRequest.Context, err = expander.ExpandForApp(app, callRequest.Expand)
		if err != nil {
			return err
		}
		callRequest.Context.Subject = subj

		up, err := p.upstreamForApp(app)
		if err != nil {
			return err
		}
		return upstream.Notify(up, callRequest)
	}

	for _, sub := range subs {
		err := notify(sub)
		if err != nil {
			// TODO log err
			continue
		}
	}
	return nil
}

func (p *Proxy) upstreamForApp(app *apps.App) (api.Upstream, error) {
	if !p.AppIsEnabled(app) {
		return nil, errors.Errorf("%s is disabled")
	}
	switch app.Type {
	case apps.AppTypeHTTP:
		return uphttp.NewUpstream(app), nil

	case apps.AppTypeAWSLambda:
		return upawslambda.NewUpstream(app, p.aws), nil

	case apps.AppTypeBuiltin:
		up := p.builtinUpstreams[app.AppID]
		if up == nil {
			return nil, errors.Errorf("builtin app not found: %s", app.AppID)
		}
		return up, nil

	default:
		return nil, errors.Errorf("not a valid app type: %s", app.Type)
	}
}

func (p *Proxy) AddBuiltinUpstream(appID apps.AppID, up api.Upstream) {
	if p.builtinUpstreams == nil {
		p.builtinUpstreams = map[apps.AppID]api.Upstream{}
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
