// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (p *Proxy) UninstallApp(r *incoming.Request, cc apps.Context, appID apps.AppID) (string, error) {
	mm := p.conf.MattermostAPI()
	app, err := p.store.App.Get(r, appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	var message string
	if app.OnUninstall != nil {
		resp := p.call(r, *app, *app.OnUninstall, &cc)
		if resp.Type == apps.CallResponseTypeError {
			r.Log.WithError(resp).Warnf("OnUninstall failed, uninstalling the app anyway")
		} else {
			message = resp.Text
		}
	}

	if message == "" {
		message = fmt.Sprintf("Uninstalled %s", app.DisplayName)
	}

	if app.MattermostOAuth2 != nil {
		// Delete oauth app.
		if err = mm.OAuth.Delete(app.MattermostOAuth2.Id); err != nil {
			return "", errors.Wrapf(err, "failed to delete Mattermost OAuth2 for %s", app.AppID)
		}

		// Only clear the store. The Mattermost Server will take care of revoking the sessions.
		if err = p.store.Session.DeleteAllForApp(r, app.AppID); err != nil {
			return "", errors.Wrapf(err, "failed to revoke sessions  for %s", app.AppID)
		}
	}

	// disable the bot account
	if _, err = mm.Bot.UpdateActive(app.BotUserID, false); err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	// delete app
	if err = p.store.App.Delete(r, app.AppID); err != nil {
		return "", errors.Wrapf(err, "can't delete app - %s", app.AppID)
	}

	// remove data
	err = p.store.AppKV.List(r, app.AppID, r.ActingUserID(), "", func(key string) error {
		return mm.KV.Delete(key)
	})
	if err != nil {
		return "", errors.Wrapf(err, "can't delete app data - %s", app.AppID)
	}

	r.Log.Infof("Uninstalled app.")

	p.conf.Telemetry().TrackUninstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	return message, nil
}
