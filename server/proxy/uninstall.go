// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (p *Proxy) UninstallApp(appID apps.AppID, sessionToken apps.SessionToken, actingUserID string) error {
	err := utils.EnsureSysadmin(p.mm, actingUserID)
	if err != nil {
		return err
	}

	app, err := p.store.App.Get(appID)
	if err != nil {
		return errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if app.OnUninstall != nil {
		creq := &apps.CallRequest{
			Call: *app.OnUninstall,
		}
		resp := p.Call(sessionToken, creq)
		if resp.Type == apps.CallResponseTypeError {
			return errors.Wrapf(resp, "call %s failed", creq.Path)
		}
	}

	// delete oauth app
	conf := p.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	client.SetToken(string(sessionToken))

	if app.OAuth2ClientID != "" {
		success, response := client.DeleteOAuthApp(app.OAuth2ClientID)
		if !success {
			if response.Error != nil {
				return errors.Wrapf(response.Error, "failed to delete OAuth2 App - %s", app.AppID)
			}
			return errors.Errorf("failed to delete OAuth2 App - returned with status code %d", response.StatusCode)
		}
	}

	// delete the bot account
	if err := p.mm.Bot.DeletePermanently(app.BotUserID); err != nil {
		return errors.Wrapf(err, "can't delete bot account for App - %s", app.AppID)
	}

	// delete app from proxy plugin, not removing the data
	if err := p.store.App.Delete(app.AppID); err != nil {
		return errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	p.mm.Log.Info("Uninstalled the app", "app_id", app.AppID)

	p.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: actingUserID})
	return nil
}
