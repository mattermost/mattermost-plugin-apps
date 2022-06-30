// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugOAuthConfigViewBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "view",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.oauth.config.view.label",
			Other: "view",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.oauth.config.view.description",
			Other: "View the OAuth configuration of a app.",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pDebugOAuthConfigView),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) debugOAuthConfigView(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))

	app, err := a.proxy.GetInstalledApp(appID, true)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewTextResponse(utils.JSONBlock(app.MattermostOAuth2))
}
