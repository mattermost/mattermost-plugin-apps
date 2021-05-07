// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (p *Proxy) UninstallApp(sessionID, actingUserID string, appID apps.AppID) error {
	err := utils.EnsureSysAdmin(p.mm, actingUserID)
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
		resp := p.Call(sessionID, actingUserID, creq)
		if resp.Type == apps.CallResponseTypeError {
			return errors.Wrapf(resp, "call %s failed", creq.Path)
		}
	}

	// delete oauth app
	session, err := utils.LoadSession(p.mm, sessionID, actingUserID)
	if err != nil {
		return err
	}
	conf := p.conf.GetConfig()
	asAdmin := model.NewAPIv4Client(conf.MattermostSiteURL)
	asAdmin.SetToken(session.Token)

	if app.MattermostOAuth2.ClientID != "" {
		success, response := asAdmin.DeleteOAuthApp(app.MattermostOAuth2.ClientID)
		if !success {
			if response.Error != nil {
				return errors.Wrapf(response.Error, "failed to delete Mattermost OAuth2 for %s", app.AppID)
			}
			return errors.Errorf("failed to delete Mattermost OAuth2 App - returned with status code %d", response.StatusCode)
		}
	}

	// disable the bot account
	if app.BotAccessTokenID != "" {
		success, response := asAdmin.RevokeUserAccessToken(app.BotAccessTokenID)
		if !success {
			if response.Error != nil {
				return errors.Wrapf(response.Error, "failed to revoke bot access token for %s", app.AppID)
			}
			return errors.Errorf("failed to revoke bot access token for %s, returned with status code %d", app.AppID, response.StatusCode)
		}
	}

	_, err = p.mm.Bot.UpdateActive(app.BotUserID, false)
	if err != nil {
		return errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	// delete app from proxy plugin, not removing the data
	if err := p.store.App.Delete(app.AppID); err != nil {
		return errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	p.mm.Log.Info("Uninstalled the app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(actingUserID)
	return nil
}
