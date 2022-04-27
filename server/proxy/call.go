// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

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

func NewProxyCallResponse(response apps.CallResponse) CallResponse {
	return CallResponse{
		CallResponse: response,
	}
}

func (r CallResponse) WithMetadata(metadata AppMetadataForClient) CallResponse {
	r.AppMetadata = metadata
	return r
}

func (p *Proxy) Call(r *incoming.Request, creq apps.CallRequest) CallResponse {
	if creq.Context.AppID == "" {
		return NewProxyCallResponse(apps.NewErrorResponse(
			utils.NewInvalidError("app_id is not set in Context, don't know what app to call")))
	}

	app, err := p.store.App.Get(creq.Context.AppID)
	if err != nil {
		return NewProxyCallResponse(apps.NewErrorResponse(err))
	}

	cresp, _ := p.callApp(r, *app, creq)
	return NewProxyCallResponse(cresp).WithMetadata(AppMetadataForClient{
		BotUserID:   app.BotUserID,
		BotUsername: app.BotUsername,
	})
}

func (p *Proxy) call(r *incoming.Request, app apps.App, call apps.Call, cc *apps.Context, valuePairs ...interface{}) apps.CallResponse {
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
	cresp, _ := p.callApp(r, app, apps.CallRequest{
		Call:    call,
		Context: *cc,
		Values:  values,
	})
	return cresp
}

func (p *Proxy) callApp(r *incoming.Request, app apps.App, creq apps.CallRequest) (apps.CallResponse, error) {
	respondErr := func(err error) (apps.CallResponse, error) {
		return apps.NewErrorResponse(err), err
	}

	conf := p.conf.Get()

	if !p.appIsEnabled(app) {
		return respondErr(errors.Errorf("%s is disabled", app.AppID))
	}

	if creq.Path[0] != '/' {
		return respondErr(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return respondErr(errors.Wrap(err, "failed to clean call path"))
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return respondErr(errors.Wrap(err, "failed to get upstream"))
	}

	cc := creq.Context
	cc = r.UpdateAppContext(cc)
	creq.Context, err = p.expandContext(r, app, &cc, creq.Expand)
	if err != nil {
		return respondErr(errors.Wrap(err, "failed to expand context"))
	}

	cresp, err := upstream.Call(r.Ctx(), up, app, creq)
	if err != nil {
		return cresp, errors.Wrap(err, "upstream call failed")
	}
	if cresp.Type == "" {
		cresp.Type = apps.CallResponseTypeOK
	}

	if cresp.Form != nil {
		clean, err := cleanForm(*cresp.Form, conf, app.AppID)
		if err != nil {
			r.Log.WithError(err).Debugf("invalid form")
		}
		cresp.Form = &clean
	}

	return cresp, nil
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
	_, err := p.callApp(r, app, apps.CallRequest{Call: apps.DefaultPing})

	return err == nil || errors.Cause(err) == utils.ErrNotFound
}

func (p *Proxy) timeoutRequest(r *incoming.Request, timeout time.Duration) (*incoming.Request, context.CancelFunc) {
	r = r.Clone()
	ctx, cancel := context.WithTimeout(r.Ctx(), timeout)
	incoming.WithCtx(ctx)(r)
	return r, cancel
}
