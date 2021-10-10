// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var disableCommandBinding = apps.Binding{
	Label:       "disable",
	Location:    "disable",
	Hint:        "[ App ID ]",
	Description: "Disables an App",
	Form:        appIDForm(newAdminCall(pDisable), newAdminCall(pDisableLookup)),
}

func (a *builtinApp) disableLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed && app.Enabled
	})
}

func (a *builtinApp) disable(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.DisableApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
