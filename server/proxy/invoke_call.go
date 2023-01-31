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

func isBindingPath(app *apps.App, pathForCheck string) bool {
	if app.Bindings != nil {
		return pathForCheck == app.Bindings.Path
	}
	return pathForCheck == apps.DefaultBindings.Path
}

func (p *Proxy) InvokeCall(r *incoming.Request, creq apps.CallRequest) (*apps.App, apps.CallResponse) {
	if err := r.Check(
		r.RequireActingUser,
	); err != nil {
		return nil, apps.NewErrorResponse(err)
	}

	app, err := p.getEnabledDestination(r)
	if err != nil {
		return nil, apps.NewErrorResponse(err)
	}
	if creq.Context.AppID != app.AppID {
		return nil, apps.NewErrorResponse(utils.NewInvalidError("incoming.Request validation error: app_id mismatch"))
	}

	if creq.Path[0] != '/' {
		return app, apps.NewErrorResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}

	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return app, apps.NewErrorResponse(errors.Wrap(err, "failed to clean call path"))
	}
	creq.Path = cleanPath

	err = checkForForbiddenPath(app, creq.Path)
	if err != nil {
		return app, apps.NewErrorResponse(errors.Wrap(err, "forbidden call path"))
	}

	appRequest := r.WithDestination(app.AppID)
	cresp := p.callApp(appRequest, app, creq, false)

	return app, cresp
}

// checkForForbiddenPath checks if the call path matches on of the call paths defined in the manifest, expect /bindings.
// These should only be called by the proxy directly, and not by the user.
func checkForForbiddenPath(app *apps.App, path string) error {
	matchesCallPath := func(call *apps.Call) bool {
		if call == nil {
			return false
		}

		return call.Path == path
	}

	manifest := app.Manifest

	if matchesCallPath(manifest.OnInstall) {
		return errors.Errorf("path %s defined as on_install.path", path)
	}
	if matchesCallPath(manifest.OnUninstall) {
		return errors.Errorf("path %s defined as on_uninstall.path", path)
	}
	if matchesCallPath(manifest.OnVersionChanged) {
		return errors.Errorf("path %s defined as on_version_changed.path", path)
	}
	if matchesCallPath(manifest.OnEnable) {
		return errors.Errorf("path %s defined as on_enable.path", path)
	}
	if matchesCallPath(manifest.OnDisable) {
		return errors.Errorf("path %s defined as on_disable.path", path)
	}
	if matchesCallPath(manifest.GetOAuth2ConnectURL) {
		return errors.Errorf("path %s defined as get_oauth2_connect_url.path", path)
	}
	if matchesCallPath(manifest.OnOAuth2Complete) {
		return errors.Errorf("path %s defined as on_oauth2_complete.path", path)
	}
	if matchesCallPath(manifest.OnRemoteWebhook) {
		return errors.Errorf("path %s defined as on_remote_webhook.path", path)
	}

	return nil
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

	if cresp.Type != apps.CallResponseTypeError &&
		!isBindingPath(app, creq.Call.Path) &&
		cresp.RefreshBindings && r.ActingUserID() != "" {
		p.dispatchRefreshBindingsEvent(r.ActingUserID())
	}

	return cresp
}
