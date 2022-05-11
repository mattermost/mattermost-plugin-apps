// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
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

func (p *Proxy) Call(r *incoming.Request, appID apps.AppID, creq apps.CallRequest) CallResponse {
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

	if appID == "" {
		return respondErr(utils.NewInvalidError("app_id is not set in request, don't know what app to call"))
	}
	if creq.Context.AppID != appID {
		return respondErr(utils.NewInvalidError("incoming.Request validation error: app_id mismatch"))
	}

	app, err := p.store.App.Get(appID)
	if err != nil {
		return respondErr(err)
	}

	if creq.Path[0] != '/' {
		return respondErr(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return respondErr(errors.Wrap(err, "failed to clean call path"))
	}
	creq.Path = cleanPath

	cresp := p.callApp(r.ToApp(app), creq)

	return CallResponse{
		CallResponse: cresp,
		AppMetadata: AppMetadataForClient{
			BotUserID:   app.BotUserID,
			BotUsername: app.BotUsername,
		},
	}
}

// <>/<> TODO: need to cleanup creq (Context) here? or assume it's good as is?
func (p *Proxy) call(r *incoming.Request, call apps.Call, cc *apps.Context, valuePairs ...interface{}) apps.CallResponse {
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
	cresp := p.callApp(r, apps.CallRequest{
		Call:    call,
		Context: *cc,
		Values:  values,
	})
	return cresp
}

// callApp in an internal method to execute a call to an upstream app. It does
// not perform any cleanup of the inputs.
//
// It returns the CallResponse to return to the client, and a separate error for the
func (p *Proxy) callApp(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	if r.To() == nil {
		return apps.NewErrorResponse(errors.New("internal unreachable error: no destination app in the incoming request"))
	}
	app := *r.To()
	if !p.appIsEnabled(app) {
		return apps.NewErrorResponse(errors.Errorf("app %s is disabled", app.AppID))
	}

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrapf(err, "no available upstream for %s", app.AppID))
	}

	// expand
	expanded, err := p.expandContext(r, &creq.Context, creq.Expand)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to expand context"))
	}
	creq.Context = *expanded

	cresp, err := upstream.Call(r.Ctx(), up, app, creq)
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

// normalizeStaticPath converts a given URL to a absolute one pointing to a static asset if needed.
// If icon is an absolute URL, it's not changed.
// Otherwise assume it's a path to a static asset and the static path URL prepended.
func normalizeStaticPath(conf config.Config, appID apps.AppID, icon string) (string, error) {
	if !strings.HasPrefix(icon, "http://") && !strings.HasPrefix(icon, "https://") {
		cleanIcon, err := utils.CleanStaticPath(icon)
		if err != nil {
			return "", errors.Wrap(err, "invalid icon path")
		}

		icon = conf.StaticURL(appID, cleanIcon)
	}

	return icon, nil
}

func (p *Proxy) GetStatic(r *incoming.Request, appID apps.AppID, path string) (io.ReadCloser, int, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		r.Log = r.Log.WithError(err)
		return nil, status, err
	}
	r = r.ToApp(app)

	return p.getStatic(r, *app, path)
}

func (p *Proxy) getStatic(r *incoming.Request, app apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return up.GetStatic(r.Ctx(), app, path)
}

// pingApp checks if the app is accessible. Call its ping path with nothing
// expanded, ignore 404 errors coming back and consider everything else a
// "success".
func (p *Proxy) pingApp(r *incoming.Request, app apps.App) (reachable bool) {
	if !p.appIsEnabled(app) {
		return false
	}

	up, err := p.upstreamForApp(app)
	if err != nil {
		return false
	}

	_, err = upstream.Call(r.Ctx(), up, app, apps.CallRequest{
		Call: apps.DefaultPing,
	})
	return err == nil || errors.Cause(err) == utils.ErrNotFound
}
