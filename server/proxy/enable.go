// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) EnableApp(in Incoming, cc apps.Context, appID apps.AppID) (string, error) {
	log := p.conf.Logger().With("app_id", appID)
	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}
	if !app.Disabled {
		return fmt.Sprintf("%s is already enabled", app.DisplayName), nil
	}

	asAdmin, err := p.getAdminClient(in)
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}
	_, err = asAdmin.EnableBot(app.BotUserID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to enable bot account for %s", app.AppID)
	}

	// Enable the app in the store first to allow calls to it
	app.Disabled = false
	err = p.store.App.Save(*app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to save app. appID: %s", appID)
	}

	var message string
	if app.OnEnable != nil {
		resp := p.callApp(in, *app, apps.CallRequest{
			Call:    *app.OnEnable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			log.WithError(err).Warnf("OnEnable failed, enabling app anyway")
		} else {
			message = resp.Markdown
		}
	}
	log.Infof("Enabled app")
	p.dispatchRefreshBindingsEvent(cc.ActingUserID)

	if message == "" {
		message = fmt.Sprintf("Enabled %s", app.DisplayName)
	}
	return message, nil
}

func (p *Proxy) DisableApp(in Incoming, cc apps.Context, appID apps.AppID) (string, error) {
	log := p.conf.Logger().With("app_id", appID)
	app, err := p.GetInstalledApp(appID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}

	if app.Disabled {
		return fmt.Sprintf("%s is already disabled", app.DisplayName), nil
	}

	// Call the app first as later it's disabled
	var message string
	if app.OnDisable != nil {
		resp := p.callApp(in, *app, apps.CallRequest{
			Call:    *app.OnDisable,
			Context: cc,
		})
		if resp.Type == apps.CallResponseTypeError {
			log.WithError(err).Warnf("OnDisable failed, disabling app anyway")
		} else {
			message = resp.Markdown
		}
	}

	if message == "" {
		message = fmt.Sprintf("Disabled %s", app.DisplayName)
	}

	asAdmin, err := p.getAdminClient(in)
	if err != nil {
		return "", errors.Wrap(err, "failed to get an admin HTTP client")
	}
	_, err = asAdmin.DisableBot(app.BotUserID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to disable bot account for %s", app.AppID)
	}

	app.Disabled = true
	err = p.store.App.Save(*app)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get app. appID: %s", appID)
	}
	log.Infof("Disabled app")
	p.dispatchRefreshBindingsEvent(in.ActingUserID)

	return message, nil
}

func (p *Proxy) appIsEnabled(app apps.App) bool {
	if app.AppType == apps.AppTypeBuiltin {
		return true
	}
	if app.Disabled {
		return false
	}
	if m, _ := p.store.Manifest.Get(app.AppID); m == nil {
		return false
	}
	return true
}
