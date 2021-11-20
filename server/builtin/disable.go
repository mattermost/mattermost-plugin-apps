// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *builtinApp) disableCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.disable.label",
			Other: "disable",
		}),
		Location: "disable",
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.disable.hint",
			Other: "[ App ID ]",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.disable.description",
			Other: "Disables an App",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pDisable),
			Fields: []apps.Field{
				a.appIDField(LookupEnabledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) disable(creq apps.CallRequest) apps.CallResponse {
	out, err := a.proxy.DisableApp(
		proxy.NewIncomingFromContext(creq.Context),
		creq.Context,
		apps.AppID(creq.GetValue(fAppID, "")))
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(out)
}
