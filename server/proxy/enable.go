// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

func (p *Proxy) EnableApp(sessionID, actingUserID string, cc *apps.Context, appID apps.AppID) (md.MD, error) {
	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if !app.Disabled {
		return md.MD(fmt.Sprintf("%s is already enabled", app.DisplayName)), nil
	}

	conf := p.conf.GetConfig()
	asAdmin, err := utils.ClientFromSession(p.mm, conf.MattermostSiteURL, sessionID, actingUserID)
	if err != nil {
		return "", err
	}

	_, response := asAdmin.EnableBot(app.BotUserID)
	if response.Error != nil {
		return "", errors.Wrapf(response.Error, "failed to enable bot account for %s", app.AppID)
	}

	// Enable the app in the store first to allow calls to it
	app.Disabled = false
	err = p.store.App.Save(app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to save app. appID: %s", appID)
	}

	var message md.MD
	if app.OnEnable != nil {
		resp := p.Call(sessionID, actingUserID, &apps.CallRequest{
			Call:    *app.OnEnable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnEnable failed, enabling app anyway", "err", resp.Error(), "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Enabled %s", app.DisplayName))
	}

	p.mm.Log.Info("Enabled an app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(actingUserID)

	return message, nil
}

func (p *Proxy) DisableApp(sessionID, actingUserID string, cc *apps.Context, appID apps.AppID) (md.MD, error) {
	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if app.Disabled {
		return md.MD(fmt.Sprintf("%s is already disabled", app.DisplayName)), nil
	}

	// Call the app first as later it's disabled
	var message md.MD
	if app.OnDisable != nil {
		resp := p.Call(sessionID, actingUserID, &apps.CallRequest{
			Call:    *app.OnDisable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnDisable failed, disabling app anyway", "err", resp.Error(), "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Disabled %s", app.DisplayName))
	}

	conf := p.conf.GetConfig()
	asAdmin, err := utils.ClientFromSession(p.mm, conf.MattermostSiteURL, sessionID, actingUserID)
	if err != nil {
		return "", err
	}

	// disable app, not removing the data
	_, response := asAdmin.DisableBot(app.BotUserID)
	if response.Error != nil {
		return "", errors.Wrapf(response.Error, "failed to disable bot account for %s", app.AppID)
	}

	app.Disabled = true
	err = p.store.App.Save(app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	p.mm.Log.Info("Disabled an app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(actingUserID)

	return message, nil
}

func (p *Proxy) AppIsEnabled(app *apps.App) bool {
	if app.AppType == apps.AppTypeBuiltin {
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
