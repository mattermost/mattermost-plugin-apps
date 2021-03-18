// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) UninstallApp(appID apps.AppID) error {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	creq := &apps.CallRequest{
		Call: *app.OnUninstall,
	}
	resp := p.Call("", creq)
	if resp.Type == apps.CallResponseTypeError {
		return errors.Wrapf(resp, "call %s failed", creq.Path)
	}

	// delete oauth app
	conf := p.conf.GetConfig()
	client := model.NewAPIv4Client(conf.MattermostSiteURL)
	// client.SetToken(string(adminToken))

	if app.OAuth2ClientID != "" {
		success, response := client.DeleteOAuthApp(app.OAuth2ClientID)
		if !success || response.StatusCode != http.StatusNoContent {
			return errors.Wrapf(response.Error, "failed to delete OAuth2 App - %s", app.AppID)
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

	return nil
}
