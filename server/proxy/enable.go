// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
)

func (p *Proxy) EnableApp(c *request.Context, cc apps.Context, appID apps.AppID) (string, error) {
	c.SetAppID(appID)

	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}
	if !app.Disabled {
		return fmt.Sprintf("%s is already enabled", app.DisplayName), nil
	}

	asAdmin, err := c.GetMMClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}
	_, err = asAdmin.EnableBot(app.BotUserID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to enable bot account for %s", app.AppID)
	}

	// Enable the app in the store first to allow calls to it
	app.Disabled = false
	err = p.store.App.Save(*app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to save app. appID: %s", appID)
	}

	var message string
	if app.OnEnable != nil {
		resp := p.call(c, *app, *app.OnEnable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			c.Log.WithError(err).Warnf("OnEnable failed, enabling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	c.Log.Infof("Enabled app")

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	if message == "" {
		message = fmt.Sprintf("Enabled %s", app.DisplayName)
	}
	return message, nil
}

func (p *Proxy) DisableApp(c *request.Context, cc apps.Context, appID apps.AppID) (string, error) {
	c.SetAppID(appID)

	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if app.Disabled {
		return fmt.Sprintf("%s is already disabled", app.DisplayName), nil
	}

	// Call the app first as later it's disabled
	var message string
	if app.OnDisable != nil {
		resp := p.call(c, *app, *app.OnDisable, &cc)
		if resp.Type == apps.CallResponseTypeError {
			c.Log.WithError(err).Warnf("OnDisable failed, disabling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Disabled %s", app.DisplayName)
	}

	asAdmin, err := c.GetMMClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}
	_, err = asAdmin.DisableBot(app.BotUserID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	app.Disabled = true
	err = p.store.App.Save(*app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	c.Log.Infof("Disabled app")

	p.dispatchRefreshBindingsEvent(c.ActingUserID())

	return message, nil
}

func (p *Proxy) appIsEnabled(app apps.App) bool {
	if app.DeployType == apps.DeployBuiltin {
		return true
	}
	if app.Disabled {
		return false
	}
	if m, _ := p.store.Manifest.Get(app.AppID); m == nil {
		return false
	}
	return true
}
