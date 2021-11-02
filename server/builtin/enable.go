// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var enableCall = apps.Call{
	Path: pEnable,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) enable() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "enable",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.label",
					Other: "enable",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.hint",
					Other: "[ App ID ]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.enable.description",
					Other: "Enable an App",
				}),
				Call: &enableCall,
				Form: a.appIDForm(enableCall, loc),
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
