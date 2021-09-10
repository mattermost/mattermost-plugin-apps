// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var disableCall = apps.Call{
	Path: pDisable,
	Expand: &apps.Expand{
		AdminAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) disableCommandBinding() apps.Binding {
	return apps.Binding{
		Label:       "disable",
		Location:    "disable",
		Hint:        "[ App ID ]",
		Description: "Disables an App",
		Call:        &disableCall,
		Form:        appIDForm(disableCall),
	}
}

func (a *builtinApp) disableLookup(creq apps.CallRequest) ([]apps.SelectOption, error) {
	return a.lookupAppID(creq, func(app apps.ListedApp) bool {
		return app.Installed && app.Enabled
	})
}

func (a *builtinApp) disableSubmit(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.DisableApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return mdResponse(out)
}
