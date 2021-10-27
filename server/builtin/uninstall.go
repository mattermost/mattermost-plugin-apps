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

func (a *builtinApp) uninstall() handler {
	return handler{
		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "uninstall",
				Location:    "uninstall",
				Hint:        "[ App ID ]",
				Description: "Uninstall an App.",
				Call:        &uninstallCall,
				Form:        appIDForm(uninstallCall),
			}
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed
			})
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			out, err := a.proxy.UninstallApp(
				proxy.NewIncomingFromContext(creq.Context),
				creq.Context,
				apps.AppID(creq.GetValue(fAppID, "")))
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			return apps.NewTextResponse(out)
		},
	}
}
