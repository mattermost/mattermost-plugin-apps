// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var debugBindingsCall = apps.Call{
	Path: pDebugBindings,
	Expand: &apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	},
}

func (a *builtinApp) debugBindings() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			form := a.appIDForm(debugBindingsCall, loc)
			if len(form.Fields) > 0 && form.Fields[0].Name == fAppID {
				form.Fields[0].IsRequired = false
			}

			return apps.Binding{
				Location: "bindings",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.bindings.label",
					Other: "bindings",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.bindings.description",
					Other: "Display all bindings for the current context",
				}),
				Call: &debugBindingsCall,
				Form: form,
			}
		},

		lookupf: func(r *incoming.Request, creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(r, creq, func(app apps.ListedApp) bool {
				return app.Installed && app.Enabled
			})
		},

		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			var bindings []apps.Binding
			if appID == "" {
				var err error
				bindings, err = a.proxy.GetBindings(r, creq.Context)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
			} else {
				r.SetAppID(appID)

				app, err := a.proxy.GetInstalledApp(r, appID)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				bindings = a.proxy.GetAppBindings(r, creq.Context, *app)
			}
			return apps.NewTextResponse(utils.JSONBlock(bindings))
		},
	}
}
