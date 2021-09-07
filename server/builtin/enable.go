// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var enableCall = apps.Call{
	Path: pEnable,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) enableCommandBinding() apps.Binding {
	return apps.Binding{
		Label:       "enable",
		Location:    "enable",
		Hint:        "[ App ID ]",
		Description: "Enables an App",
		Call:        &enableCall,
		Form:        appIDForm(enableCall),
	}
}

func (a *builtinApp) enableLookup(creq apps.CallRequest) ([]apps.SelectOption, error) {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed && !app.Enabled
	})
}

func (a *builtinApp) enableSubmit(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.EnableApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return mdResponse(out)
}
