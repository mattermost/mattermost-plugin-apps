// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream/upawslambda"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (p *Proxy) Call(debugSessionToken apps.SessionToken, c *apps.CallRequest) *apps.CallResponse {
	app, err := p.store.App.Get(c.Context.AppID)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	cc := p.conf.GetConfig().SetContextDefaultsForApp(c.Context, c.Context.AppID)

	expander := p.newExpander(cc, p.mm, p.conf, p.store, debugSessionToken)
	cc, err = expander.ExpandForApp(app, c.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	clone := *c
	clone.Context = cc

	return upstream.Call(up, &clone)
}

func (p *Proxy) Notify(cc *apps.Context, subj apps.Subject) error {
	subs, err := p.store.Subscription.Get(subj, cc.TeamID, cc.ChannelID)
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
		app, err := p.store.App.Get(sub.AppID)
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

func (p *Proxy) GetAsset(appID apps.AppID, path string) (io.ReadCloser, int, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Cause(err) == utils.ErrNotFound {
			status = http.StatusNotFound
		}
		return nil, status, err
	}
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(path)
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
	if !p.AppIsEnabled(app) {
		return nil, errors.Errorf("%s is disabled", app.AppID)
	}
	switch app.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewUpstream(app), nil

	case apps.AppTypeAWSLambda:
		return upawslambda.NewUpstream(app, p.aws, p.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		up := p.builtinUpstreams[app.AppID]
		if up == nil {
			return nil, errors.Errorf("builtin app not found: %s", app.AppID)
		}
		return up, nil

	default:
		return nil, errors.Errorf("not a valid app type: %s", app.AppType)
	}
}
