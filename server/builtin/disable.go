// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var disableCall = apps.Call{
	Path: pDisable,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) disable() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "disable",
				Label:       a.conf.Local(loc, "command.disable.label"),
				Description: a.conf.Local(loc, "command.disable.description"),
				Hint:        a.conf.Local(loc, "command.disable.hint"),
				Call:        &disableCall,
				Form:        a.appIDForm(disableCall, loc),
			}
		},

		// Lookup returns the list of eligible Apps.
		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, func(app apps.ListedApp) bool {
				return app.Installed && app.Enabled
			})
		},

		// Submit disables an app.
		submitf: func(creq apps.CallRequest) apps.CallResponse {
			out, err := a.proxy.DisableApp(
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
