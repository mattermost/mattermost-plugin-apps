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

func (a *builtinApp) enable() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func() apps.Binding {
			return apps.Binding{
				Label:       "enable",
				Location:    "enable",
				Hint:        "[ App ID ]",
				Description: "Enable an App.",
				Call:        &enableCall,
				Form:        appIDForm(enableCall),
			}
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed && !app.Enabled
			})
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			out, err := a.proxy.EnableApp(
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
