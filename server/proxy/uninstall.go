// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
)

func (p *Proxy) UninstallApp(client MMClient, sessionID string, cc *apps.Context, appID apps.AppID) (md.MD, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	var message md.MD
	if app.OnUninstall != nil {
		creq := &apps.CallRequest{
			Call:    *app.OnUninstall,
			Context: cc,
		}
		resp := p.Call(sessionID, cc.ActingUserID, creq)
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnUninstall failed, uninstalling app anyway", "err", resp.Error(), "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Uninstalled %s", app.DisplayName))
	}

	// delete oauth app
	if app.MattermostOAuth2.ClientID != "" {
		if err = client.DeleteOAuthApp(app.MattermostOAuth2.ClientID); err != nil {
			return "", errors.Wrapf(err, "failed to delete Mattermost OAuth2 for %s", app.AppID)
		}
	}

	// revoke bot account token if there is one
	if app.BotAccessTokenID != "" {
		if err = client.RevokeUserAccessToken(app.BotAccessTokenID); err != nil {
			return "", errors.Wrapf(err, "failed to revoke bot access token for %s", app.AppID)
		}
	}

	// disable the bot account
	if _, err = client.DisableBot(app.BotUserID); err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	// delete app
	if err = p.store.App.Delete(app.AppID); err != nil {
		return "", errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	// in on-prem mode the manifest need to be deleted as every install add a manifest anyway
	conf := p.conf.GetConfig()
	if !conf.MattermostCloudMode {
		if err = p.store.Manifest.DeleteLocal(app.AppID); err != nil {
			return "", errors.Wrapf(err, "can't delete manifest for uninstalled app - %s", app.AppID)
		}
	}

	// remove data
	if err = p.store.AppKV.DeleteAll(app.BotUserID); err != nil {
		return "", errors.Wrapf(err, "can't delete app data - %s", app.AppID)
	}

	p.mm.Log.Info("Uninstalled an app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	return message, nil
}
