// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var uninstallCall = apps.Call{
	Path: pUninstall,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) uninstallCommandBinding() apps.Binding {
	return apps.Binding{
		Label:       "uninstall",
		Location:    "uninstall",
		Hint:        "[ App ID ]",
		Description: "Uninstalls an App",
		Call:        &uninstallCall,
	}
}

func (a *builtinApp) uninstallForm(creq apps.CallRequest) apps.CallResponse {
	return appIDForm(uninstallCall)
}

func (a *builtinApp) uninstallLookup(creq apps.CallRequest) apps.CallResponse {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed
	})
}

func (a *builtinApp) uninstallSubmit(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.UninstallApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return mdResponse(out)
}
