// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (p *Proxy) UninstallApp(r *incoming.Request, cc apps.Context, appID apps.AppID, force bool) (string, error) {
	if err := r.Check(
		r.RequireActingUser,
		r.RequireSysadminOrPlugin,
	); err != nil {
		return "", err
	}

	mm := p.conf.MattermostAPI()
	app, err := p.store.App.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app, appID: %s", appID)
	}

	// Disable the app and clean out all of its sessions.
	_, err = p.DisableApp(r, cc, appID)
	if err != nil {
		return "", errors.Wrap(err, "failed to disable app")
	}

	message := fmt.Sprintf("Uninstalled %s", app.DisplayName)
	if app.OnUninstall != nil {
		resp := p.call(r, app, *app.OnUninstall, &cc)
		switch resp.Type {
		case apps.CallResponseTypeError:
			if !force {
				return "", errors.Wrap(resp, "app canceled uninstall request")
			}
			message = fmt.Sprintf("Force-uninstalled %s, despite error: %s", app.DisplayName, resp.Text)

		case apps.CallResponseTypeOK:
			message = fmt.Sprintf("Uninstalled %s, with message: %s", app.DisplayName, resp.Text)
		}
	}

	// Delete OAuth app.
	if app.MattermostOAuth2 != nil {
		if err = mm.OAuth.Delete(app.MattermostOAuth2.Id); err != nil {
			return "", errors.Wrapf(err, "failed to delete Mattermost OAuth2 for %s, the app is left disabled", appID)
		}
	}

	// Disable the app's bot account.
	if _, err = mm.Bot.UpdateActive(app.BotUserID, false); err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s, the app is left disabled", appID)
	}

	// Remove the app's KV store data.
	err = p.store.AppKV.List(r.WithSourceAppID(appID), "", func(key string) error {
		return mm.KV.Delete(key)
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to clear KV data for %s, the app is left disabled", appID)
	}

	// TODO: delete OAuth2 user objects

	if err = p.appservices.UnsubscribeApp(r, appID); err != nil {
		return "", errors.Wrapf(err, "failed to clear subscriptions for %s, the app is left disabled", appID)
	}

	// Delete the main record of the app.
	if err = p.store.App.Delete(r, app.AppID); err != nil {
		return "", errors.Wrapf(err, "can't delete app %s, the app is left disabled", appID)
	}

	r.Log.Infof("Uninstalled app.")

	p.conf.Telemetry().TrackUninstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	return message, nil
}
