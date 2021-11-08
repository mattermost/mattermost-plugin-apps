// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

var uninstallCall = apps.Call{
	Path: pUninstall,
	Expand: &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) uninstall() handler {
	return handler{
		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "uninstall",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.uninstall.label",
					Other: "uninstall",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.uninstall.hint",
					Other: "[ App ID ]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.uninstall.description",
					Other: "Uninstall an App",
				}),
				Call: &uninstallCall,
				Form: a.appIDForm(uninstallCall, loc),
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
