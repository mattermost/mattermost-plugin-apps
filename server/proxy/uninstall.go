// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
)

func (p *Proxy) UninstallApp(c *request.Context, cc apps.Context, appID apps.AppID) (string, error) {
	mm := p.conf.MattermostAPI()
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	var message string
	if app.OnUninstall != nil {
		resp := p.call(c, *app, *app.OnUninstall, &cc)
		if resp.Type == apps.CallResponseTypeError {
			c.Log.WithError(resp).Warnf("OnUninstall failed, uninstalling the app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Uninstalled %s", app.DisplayName)
	}

	asAdmin, err := c.GetMMClient()
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}

	// delete oauth app
	c.Log.Errorf("app.MattermostOAuth2: %#+v\n", app.MattermostOAuth2)
	if app.MattermostOAuth2 != nil {
		if err = asAdmin.DeleteOAuthApp(app.MattermostOAuth2.Id); err != nil {
			return "", errors.Wrapf(err, "failed to delete Mattermost OAuth2 for %s", app.AppID)
		}

		if err = p.sessionService.RevokeSessionsForApp(c, app.AppID); err != nil {
			return "", errors.Wrapf(err, "failed to revoke sessions  for %s", app.AppID)
		}
	}

	// revoke bot account token if there is one
	if app.BotAccessTokenID != "" {
		if err = asAdmin.RevokeUserAccessToken(app.BotAccessTokenID); err != nil {
			return "", errors.Wrapf(err, "failed to revoke bot access token for %s", app.AppID)
		}
	}

	// disable the bot account
	if _, err = asAdmin.DisableBot(app.BotUserID); err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	// delete app
	if err = p.store.App.Delete(app.AppID); err != nil {
		return "", errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	// remove data
	err = p.store.AppKV.List(app.BotUserID, "", func(key string) error {
		return mm.KV.Delete(key)
	})
	if err != nil {
		return "", errors.Wrapf(err, "can't delete app data - %s", app.AppID)
	}

	c.Log.Infof("Uninstalled app.")

	p.conf.Telemetry().TrackUninstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(c.ActingUserID())

	return message, nil
}
