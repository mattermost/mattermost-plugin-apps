// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// CallResponse contains everything the CallResponse struct contains, plus some additional
// data for the client, such as information about the App's bot account.
//
// Apps will use the CallResponse struct to respond to a CallRequest, and the proxy will
// decorate the response using the CallResponse to provide additional information.
type CallResponse struct {
	apps.CallResponse

	// Used to provide info about the App to client, e.g. the bot user id
	AppMetadata AppMetadataForClient `json:"app_metadata"`
}

type AppMetadataForClient struct {
	BotUserID   string `json:"bot_user_id,omitempty"`
	BotUsername string `json:"bot_username,omitempty"`
}

func (p *Proxy) InvokeCall(r *incoming.Request, creq apps.CallRequest) CallResponse {
	var app *apps.App
	respondErr := func(err error) CallResponse {
		out := CallResponse{
			CallResponse: apps.NewErrorResponse(err),
		}
		if app != nil {
			out.AppMetadata = AppMetadataForClient{
				BotUserID:   app.BotUserID,
				BotUsername: app.BotUsername,
			}
		}
		return out
	}

	if err := r.Check(
		r.RequireActingUser,
	); err != nil {
		return respondErr(err)
	}

	app, err := p.getEnabledDestination(r)
	if err != nil {
		return respondErr(err)
	}
	if creq.Context.AppID != app.AppID {
		return respondErr(utils.NewInvalidError("incoming.Request validation error: app_id mismatch"))
	}

	if creq.Path[0] != '/' {
		return respondErr(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return respondErr(errors.Wrap(err, "failed to clean call path"))
	}
	creq.Path = cleanPath

	appRequest := r.WithDestination(app.AppID)
	cresp := p.callApp(appRequest, app, creq, false)

	return CallResponse{
		CallResponse: cresp,
		AppMetadata: AppMetadataForClient{
			BotUserID:   app.BotUserID,
			BotUsername: app.BotUsername,
		},
	}
}

// <>/<> TODO: need to cleanup creq (Context) here? or assume it's good as is?
func (p *Proxy) call(r *incoming.Request, app *apps.App, call apps.Call, cc *apps.Context, valuePairs ...interface{}) apps.CallResponse {
	values := map[string]interface{}{}
	for len(valuePairs) > 0 {
		if len(valuePairs) == 1 {
			return apps.NewErrorResponse(
				errors.Errorf("mismatched parameter count, no value for %v", valuePairs[0]))
		}
		key, ok := valuePairs[0].(string)
		if !ok {
			return apps.NewErrorResponse(
				errors.Errorf("mismatched type %T for key %v, expected string", valuePairs[0], valuePairs[0]))
		}
		values[key] = valuePairs[1]
		valuePairs = valuePairs[2:]
	}

	if cc == nil {
		cc = &apps.Context{}
	}
	cresp := p.callApp(r, app, apps.CallRequest{
		Call:    call,
		Context: *cc,
		Values:  values,
	}, false)
	return cresp
}

// callApp in an internal method to execute a call to an upstream app. It does
// not perform any cleanup of the inputs.
func (p *Proxy) callApp(r *incoming.Request, app *apps.App, creq apps.CallRequest, notify bool) apps.CallResponse {
	// this may be invoked from various places in the code, and the Destination
	// may or may not be set in the request. Since we have the app explicitly
	// here, make sure it's set in the request
	r = r.WithDestination(app.AppID)

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrapf(err, "no available upstream for %s", app.AppID))
	}

	// expand
	expanded, err := p.expandContext(r, app, &creq.Context, creq.Expand)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to expand context"))
	}
	creq.Context = *expanded

	if notify {
		err = upstream.Notify(r.Ctx(), up, *app, creq)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "upstream call failed"))
		}
		return apps.NewTextResponse("OK")
	}

	cresp, err := upstream.Call(r.Ctx(), up, *app, creq)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "upstream call failed"))
	}
	if cresp.Type == "" {
		cresp.Type = apps.CallResponseTypeOK
	}
	if cresp.Form != nil {
		clean, err := cleanForm(*cresp.Form, p.conf.Get(), app.AppID)
		if err != nil {
			r.Log.WithError(err).Debugf("invalid form in call response")
		}
		cresp.Form = &clean
	}
	return cresp
}
