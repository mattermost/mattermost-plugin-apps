// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) uninstallCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.uninstall.label",
			Other: "uninstall",
		}),
		Location: "uninstall",
		Hint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.uninstall.hint",
			Other: "[ App ID ]",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.uninstall.description",
			Other: "Uninstall an App",
		}),

		Form: &apps.Form{
			Submit: newUserCall(pUninstall),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
				{
					Name: fForce,
					Type: apps.FieldTypeBool,
					Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.force.description",
						Other: "Forcefully uninstall the app, even if there is an error",
					}),
					Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.foce.label",
						Other: "force",
					}),
				},
			},
		},
	}
}

func (a *builtinApp) uninstall(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	force := creq.BoolValue(fForce)
	out, err := a.proxy.UninstallApp(r, creq.Context, appID, force)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
