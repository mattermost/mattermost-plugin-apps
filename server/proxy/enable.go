// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (p *Proxy) EnableApp(r *incoming.Request, cc apps.Context, appID apps.AppID) (string, error) {
	app, err := p.GetInstalledApp(r, appID)
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

	var message string
	if app.OnEnable != nil {
		resp := p.call(r, *app, *app.OnEnable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			r.Log.WithError(err).Warnf("OnEnable failed, enabling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	r.Log.Infof("Enabled app")

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	if message == "" {
		message = fmt.Sprintf("Enabled %s", app.DisplayName)
	}
	return message, nil
}

func (p *Proxy) DisableApp(r *incoming.Request, cc apps.Context, appID apps.AppID) (string, error) {
	app, err := p.GetInstalledApp(r, appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if app.Disabled {
		return fmt.Sprintf("%s is already disabled", app.DisplayName), nil
	}

	// Call the app first as later it's disabled
	var message string
	if app.OnDisable != nil {
		resp := p.call(r, *app, *app.OnDisable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			r.Log.WithError(err).Warnf("OnDisable failed, disabling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Disabled %s", app.DisplayName)
	}

	_, err = p.conf.MattermostAPI().Bot.UpdateActive(app.BotUserID, false)
	if err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	//  Only clear the store. Existing session will still work until they expire. https://mattermost.atlassian.net/browse/MM-40012
	if err = p.store.Session.DeleteAllForApp(r, app.AppID); err != nil {
		return "", errors.Wrapf(err, "failed to revoke sessions  for %s", app.AppID)
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

func (p *Proxy) appIsEnabled(r *incoming.Request, app apps.App) bool {
	if app.DeployType == apps.DeployBuiltin {
		return true
	}
	if app.Disabled {
		return false
	}
	if m, _ := p.store.Manifest.Get(r, app.AppID); m == nil {
		return false
	}
	return true
}
