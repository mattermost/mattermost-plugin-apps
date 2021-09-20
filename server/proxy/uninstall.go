// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) UninstallApp(in Incoming, cc apps.Context, appID apps.AppID) (string, error) {
	conf, _, log := p.conf.Basic()
	log = log.With("app_id", appID)
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	var message string
	if app.OnUninstall != nil {
		resp := p.callApp(in, *app, apps.CallRequest{
			Call:    *app.OnUninstall,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			log.WithError(err).Warnf("OnUninstall failed, uninstalling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Uninstalled %s", app.DisplayName)
	}

	asAdmin, err := p.getAdminClient(in)
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}
	// delete oauth app
	if app.MattermostOAuth2.ClientID != "" {
		if err = asAdmin.DeleteOAuthApp(app.MattermostOAuth2.ClientID); err != nil {
			return "", errors.Wrapf(err, "failed to delete Mattermost OAuth2 for %s", app.AppID)
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

	// in on-prem mode the manifest need to be deleted as every install add a manifest anyway
	if !conf.MattermostCloudMode {
		if err = p.store.Manifest.DeleteLocal(app.AppID); err != nil {
			return "", errors.Wrapf(err, "can't delete manifest for uninstalled app - %s", app.AppID)
		}
	}

	// remove data
	if err = p.store.AppKV.DeleteAll(app.BotUserID); err != nil {
		return "", errors.Wrapf(err, "can't delete app data - %s", app.AppID)
	}

	log.Infof("Uninstalled app.")

	p.conf.Telemetry().TrackUninstall(string(app.AppID), string(app.AppType))

	p.dispatchRefreshBindingsEvent(in.ActingUserID)

	return message, nil
}
