// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import "github.com/mattermost/mattermost-plugin-apps/server/api"

func (p *Proxy) Call(debugSessionToken api.SessionToken, c *api.Call) (*api.CallResponse, error) {
	app, err := p.store.LoadApp(c.Context.AppID)
	if err != nil {
		return nil, err
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, err
	}

	expander := p.newExpander(c.Context, p.mm, p.conf, p.store, debugSessionToken)
	expander.App = app
	cc, err := expander.Expand(c.Expand)
	if err != nil {
		return nil, err
	}
	clone := *c
	clone.Context = cc

	return up.InvokeCall(&clone)
}

func (p *Proxy) Notify(cc *api.Context, subj api.Subject) error {
	app, err := p.store.LoadApp(cc.AppID)
	if err != nil {
		return err
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return err
	}

	subs, err := p.store.LoadSubs(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	expander := p.newExpander(cc, p.mm, p.conf, p.store, "")
	expander.App = app
	for _, sub := range subs {
		req := api.Notification{
			Subject: subj,
			Context: &api.Context{},
		}
		req.Context, err = expander.Expand(sub.Expand)
		if err != nil {
			return err
		}

		// Always set the AppID for routing the request to the App
		req.Context.AppID = sub.AppID

		go func() {
			_ = up.InvokeNotification(&req)
		}()
	}
	return nil
}
