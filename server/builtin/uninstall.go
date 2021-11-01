// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var uninstallCall = apps.Call{
	Path: pUninstall,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) uninstall() handler {
	return handler{
		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "uninstall",
				Label:       a.conf.Local(loc, "command.uninstall.label"),
				Description: a.conf.Local(loc, "command.uninstall.description"),
				Hint:        a.conf.Local(loc, "command.uninstall.hint"),
				Call:        &uninstallCall,
				Form:        a.appIDForm(uninstallCall, loc),
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
