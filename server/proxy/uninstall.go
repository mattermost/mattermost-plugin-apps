// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) UninstallApp(r *incoming.Request, cc apps.Context, appID apps.AppID, force bool) (_ string, err error) {
	if err = r.Check(
		r.RequireActingUser,
		r.RequireSysadminOrPlugin,
	); err != nil {
		return "", err
	}

	mm := p.conf.MattermostAPI()
	app, err := p.appStore.Get(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app for appID: %s", appID)
	}

	u := uninstaller{
		deleteApp:                 func() error { return p.appStore.Delete(r, appID) },
		deleteAppData:             func() error { return p.appservices.DeleteAppData(r, appID, force) },
		deleteMattermostOAuth2App: func() error { return mm.OAuth.Delete(app.MattermostOAuth2.Id) },
		disableApp:                func() error { _, e := p.DisableApp(r, cc, appID); return e },
		disableBotAccount:         func() error { _, e := mm.Bot.UpdateActive(app.BotUserID, false); return e },
		revokeSessionsForApp:      func() error { return p.sessionService.RevokeSessionsForApp(r, app.AppID) },
		uninstallCall:             func() apps.CallResponse { return p.call(r, app, *app.OnUninstall, &cc) },
		log:                       r.Log,
	}

	message, err := u.uninstall(app, force)
	if err != nil {
		return "", err
	}

	r.Log.Infof("Uninstalled app %s.", appID)

	p.conf.Telemetry().TrackUninstall(string(app.AppID), string(app.DeployType))

	p.dispatchRefreshBindingsEvent(r.ActingUserID())

	return message, nil
}

type uninstaller struct {
	deleteApp                 func() error
	deleteAppData             func() error
	deleteMattermostOAuth2App func() error
	disableApp                func() error
	disableBotAccount         func() error
	revokeSessionsForApp      func() error
	uninstallCall             func() apps.CallResponse
	log                       utils.Logger
}

func (u uninstaller) uninstall(app *apps.App, force bool) (_ string, err error) {
	var errs *multierror.Error
	// Log errors on exit.
	defer func() {
		switch {
		case err != nil:
			u.log.WithError(err).Errorf("failed to uninstall app: %v", err)
		case force && errs != nil:
			u.log.Errorf("force-uninstalled app despite errors: %v", errs)
		}
	}()

	// Helper to wrap, and to log errors and continue if force is true.
	forceIgnoreErrors := func(e error, text string) error {
		if e == nil {
			return nil
		}
		if text != "" {
			e = errors.Wrap(e, string(app.AppID)+": "+text)
		}
		if !force {
			return e
		}
		errs = multierror.Append(errs, e)
		return nil
	}

	// Give the app a chance to clean up.
	var onUninstallResponse apps.CallResponse
	if app.OnUninstall != nil {
		onUninstallResponse = u.uninstallCall()
		if onUninstallResponse.Type == apps.CallResponseTypeError {
			if err = forceIgnoreErrors(onUninstallResponse, "app canceled uninstall"); err != nil {
				return "", err
			}
		}
	}

	// If the app's data cleanup fails in the middle, disable the app.
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "disabled app after failing to clean up its data")
			if disableErr := u.disableApp(); disableErr != nil {
				err = multierror.Append(&multierror.Error{},
					errors.Wrap(err, "failed to clean up app data on uninstall"),
					errors.Wrap(disableErr, "failed to disable app after failing to clean up its data"))
			}
		}
	}()

	// Only clear the session store. Existing session are revoked when the OAuth app gets deleted.
	if err = forceIgnoreErrors(u.revokeSessionsForApp(), "failed to revoke sessions"); err != nil {
		return "", err
	}

	if app.MattermostOAuth2 != nil {
		if err = forceIgnoreErrors(u.deleteMattermostOAuth2App(), "failed to delete Mattermost OAuth2 app record"); err != nil {
			return "", err
		}
	}

	if err = forceIgnoreErrors(u.disableBotAccount(), "failed to disable Mattermost bot account"); err != nil {
		return "", err
	}

	if err = forceIgnoreErrors(u.deleteAppData(), "failed to clean app's persisted data"); err != nil {
		return "", err
	}

	if err = u.deleteApp(); err != nil {
		return "", errors.Wrapf(err, "failed to delete app")
	}

	message := fmt.Sprintf("Uninstalled %s (%s)", app.AppID, app.DisplayName)
	switch {
	case force && errs != nil:
		message = fmt.Sprintf("Force-uninstalled %s (%s), despite error(s): %v", app.AppID, app.DisplayName, errs)
	case onUninstallResponse.Text != "":
		message = fmt.Sprintf("Uninstalled %s (%s), with message: %s", app.AppID, app.DisplayName, onUninstallResponse.Text)
	}

	return message, nil
}
