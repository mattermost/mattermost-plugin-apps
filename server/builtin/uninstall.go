// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) uninstallCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.uninstall.label",
			Other: "uninstall",
		}),
		Location: "uninstall",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.uninstall.hint",
			Other: "[ App ID ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "apps.command.uninstall.description",
			Other: "Uninstalls an App",
		}),
		Form: a.appIDForm(newAdminCall(pUninstall), newAdminCall(pUninstallLookup), loc),
	}
}

func (a *builtinApp) uninstallLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed
	})
}

func (a *builtinApp) uninstall(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.UninstallApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
