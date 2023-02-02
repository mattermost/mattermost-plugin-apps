// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) EnableApp(r *incoming.Request, cc apps.Context, appID apps.AppID) (string, error) {
	if err := r.Check(
		r.RequireSysadminOrPlugin,
	); err != nil {
		return "", err
	}

	app, err := p.GetInstalledApp(appID, false)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}
	if !app.Disabled {
		return fmt.Sprintf("%s is already enabled", app.DisplayName), nil
	}

	_, err = p.conf.MattermostAPI().Bot.UpdateActive(app.BotUserID, true)
	if err != nil {
		return "", errors.Wrapf(err, "failed to enable bot account for %s", app.AppID)
	}

	// Enable the app in the store first to allow calls to it
	app.Disabled = false
	err = p.store.App.Save(r, *app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to save app. appID: %s", appID)
	}

	// TODO Get this to a separate method
	oauthApi := p.conf.MattermostAPI().OAuth
	oauthApp, err := oauthApi.Get(string(appID))
	// TODO This doesn't exist yet in the MM server
	oauthApp.Disabled = false
	if err != nil {
		return "", errors.Wrapf(err, "failed to get OAuth app %s", app.AppID)
	}
	p.conf.MattermostAPI().OAuth.Update(oauthApp)
	if err != nil {
		return "", errors.Wrapf(err, "failed to update OAuth app %s", app.AppID)
	}

	var message string
	if app.OnEnable != nil {
		resp := p.call(r, app, *app.OnEnable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			r.Log.WithError(err).Warnf("OnEnable failed, enabling app anyway")
		} else {
			message = resp.Text
		}
	}

	r.Log.Infof("Enabled app")

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	if message == "" {
		message = fmt.Sprintf("Enabled %s", app.DisplayName)
	}
	return message, nil
}

func (p *Proxy) DisableApp(r *incoming.Request, cc apps.Context, appID apps.AppID) (string, error) {
	if err := r.Check(
		r.RequireSysadminOrPlugin,
	); err != nil {
		return "", err
	}

	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}
	if app.Disabled {
		return "Already disabled", nil
	}

	// Call the app first as later it's disabled
	var message string
	if app.OnDisable != nil {
		resp := p.call(r, app, *app.OnDisable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			r.Log.WithError(err).Warnf("OnDisable failed, disabling app anyway")
		} else {
			message = resp.Text
		}
	}

	if message == "" {
		message = fmt.Sprintf("Disabled %s", app.DisplayName)
	}

	_, err = p.conf.MattermostAPI().Bot.UpdateActive(app.BotUserID, false)
	if err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	// Only clear the store. Existing session will still work until they expire. https://mattermost.atlassian.net/browse/MM-40012
	if err = p.store.Session.DeleteAllForApp(r, app.AppID); err != nil {
		return "", errors.Wrapf(err, "failed to revoke sessions  for %s", app.AppID)
	}

	// TODO Get this to a separate method
	oauthApi := p.conf.MattermostAPI().OAuth
	oauthApp, err := oauthApi.Get(string(appID))
	// TODO This doesn't exist yet in the MM server
	oauthApp.Disabled = true
	if err != nil {
		return "", errors.Wrapf(err, "failed to get OAuth app %s", app.AppID)
	}
	p.conf.MattermostAPI().OAuth.Update(oauthApp)
	if err != nil {
		return "", errors.Wrapf(err, "failed to update OAuth app %s", app.AppID)
	}

	app.Disabled = true
	err = p.store.App.Save(r, *app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	r.Log.Infof("Disabled app")

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	return message, nil
}

func (p *Proxy) ensureEnabled(app *apps.App) error {
	if app.DeployType == apps.DeployBuiltin {
		// builtins can not be disabled ATM, or catch-22 with `apps`
		return nil
	}
	if app.Disabled {
		return utils.NewForbiddenError("app is disabled by the administrator: %s", app.AppID)
	}
	if m, _ := p.store.Manifest.Get(app.AppID); m == nil {
		return utils.NewForbiddenError("app is no longer listed: %s", app.AppID)
	}
	return nil
}
