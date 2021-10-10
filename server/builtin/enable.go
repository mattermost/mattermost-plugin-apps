// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var enableCommandBinding = apps.Binding{
	Label:       "enable",
	Location:    "enable",
	Hint:        "[ App ID ]",
	Description: "Enables an App",
	Form:        appIDForm(newAdminCall(pEnable), newAdminCall(pEnableLookup)),
}

func (a *builtinApp) enableLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed && !app.Enabled
	})
}

func (a *builtinApp) enable(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.EnableApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
