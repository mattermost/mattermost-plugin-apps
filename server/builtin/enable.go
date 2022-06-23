// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) enableCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.label",
			Other: "enable",
		}),
		Location: "enable",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.hint",
			Other: "[ App ID ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.enable.description",
			Other: "Enable an App",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pEnable),
			Fields: []apps.Field{
				a.appIDField(LookupDisabledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) enable(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.EnableApp(
		r,
		creq.Context,
		apps.AppID(creq.GetValue(FieldAppID, "")))
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
