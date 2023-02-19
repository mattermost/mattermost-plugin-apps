// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugSessionsViewBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "view",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.view.label",
			Other: "view",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.view.description",
			Other: "View a session.",
		}),
		Form: &apps.Form{
			Submit: newUserCall(pDebugSessionsView),
			Fields: []apps.Field{
				{
					Name: fSessionID,
					Type: apps.FieldTypeText,
					Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.session.label",
						Other: "sessionID",
					}),
					Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.session.description",
						Other: "enter the session ID",
					}),
					AutocompleteHint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "field.session.hint",
						Other: "Session ID",
					}),
					AutocompletePosition: 1,
					IsRequired:           true,
				},
			},
		},
	}
}

func (a *builtinApp) debugSessionsView(_ *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	sessionID := creq.GetValue(fSessionID, "")
	session, err := a.api.Mattermost.Session.Get(sessionID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewTextResponse(utils.JSONBlock(session))
}
