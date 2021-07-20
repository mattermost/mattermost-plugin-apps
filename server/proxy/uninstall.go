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

func (p *Proxy) UninstallApp(sessionID, actingUserID string, cc *apps.Context, appID apps.AppID) (md.MD, error) {
	err := utils.EnsureSysAdmin(p.mm, actingUserID)
	if err != nil {
		return "", err
	}

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
		resp := p.Call(sessionID, actingUserID, creq)
		if resp.Type == apps.CallResponseTypeError {
			p.mm.Log.Warn("OnUninstall failed, uninstalling app anyway", "err", resp.Error(), "app_id", app.AppID)
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = md.MD(fmt.Sprintf("Uninstalled %s", app.DisplayName))
	}

	conf := p.conf.GetConfig()
	asAdmin, err := utils.ClientFromSession(p.mm, conf.MattermostSiteURL, sessionID, actingUserID)
	if err != nil {
		return "", err
	}

	// delete oauth app
	if app.MattermostOAuth2.ClientID != "" {
		success, response := asAdmin.DeleteOAuthApp(app.MattermostOAuth2.ClientID)
		if !success {
			if response.Error != nil {
				return "", errors.Wrapf(response.Error, "failed to delete Mattermost OAuth2 for %s", app.AppID)
			}
			return "", errors.Errorf("failed to delete Mattermost OAuth2 App - returned with status code %d", response.StatusCode)
		}
	}

	// revoke bot account token if there is one
	if app.BotAccessTokenID != "" {
		success, response := asAdmin.RevokeUserAccessToken(app.BotAccessTokenID)
		if !success {
			if response.Error != nil {
				return "", errors.Wrapf(response.Error, "failed to revoke bot access token for %s", app.AppID)
			}
			return "", errors.Errorf("failed to revoke bot access token for %s, returned with status code %d", app.AppID, response.StatusCode)
		}
	}

	// disable the bot account
	_, response := asAdmin.DisableBot(app.BotUserID)
	if response.Error != nil {
		return "", errors.Wrapf(response.Error, "failed to disable bot account for %s", app.AppID)
	}

	// delete app
	if err := p.store.App.Delete(app.AppID); err != nil {
		return "", errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	// in on-prem mode the manifest need to be deleted as every install add a manifest anyway
	if !conf.MattermostCloudMode {
		if err := p.store.Manifest.DeleteLocal(app.AppID); err != nil {
			return "", errors.Wrapf(err, "can't delete manifest for uninstalled app - %s", app.AppID)
		}
	}

	// remove data
	if err := p.store.AppKV.DeleteAll(app.BotUserID); err != nil {
		return "", errors.Wrapf(err, "can't delete app data - %s", app.AppID)
	}

	p.mm.Log.Info("Uninstalled an app", "app_id", app.AppID)

	p.dispatchRefreshBindingsEvent(actingUserID)

	return message, nil
}
