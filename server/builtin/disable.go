// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

var disableCall = apps.Call{
	Path: pDisable,
	Expand: &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) disable() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "disable",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.disable.label",
					Other: "disable",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.disable.hint",
					Other: "[ App ID ]",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.disable.description",
					Other: "Disable an App",
				}),
				Call: &disableCall,
				Form: a.appIDForm(disableCall, loc),
			}
		},

		// Lookup returns the list of eligible Apps.
		lookupf: func(r *incoming.Request, creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(r, creq, func(app apps.ListedApp) bool {
				return app.Installed && app.Enabled
			})
		},

		// Submit disables an app.
		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			r.SetAppID(appID)
			out, err := a.proxy.DisableApp(
				r,
				creq.Context,
				appID,
			)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			return apps.NewTextResponse(out)
		},
	}
}
