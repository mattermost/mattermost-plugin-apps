// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var uninstallCommandBinding = apps.Binding{
	Label:       "uninstall",
	Location:    "uninstall",
	Hint:        "[ App ID ]",
	Description: "Uninstalls an App",
	Form:        appIDForm(newAdminCall(pUninstall), newAdminCall(pUninstallLookup)),
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
