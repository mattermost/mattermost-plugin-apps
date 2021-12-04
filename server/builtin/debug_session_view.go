// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var viewSessionCall = apps.Call{
	Path: pDebugSessionsView,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugSessionsView() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "view",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.view.label",
					Other: "view",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.view.description",
					Other: "View a session.",
				}),
				Call: &viewSessionCall,
				Form: &apps.Form{
					Call: &viewSessionCall,
					Fields: []apps.Field{
						{
							Name: fSessionID,
							Type: apps.FieldTypeText,
							Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.session.label",
								Other: "sessionID",
							}),
							Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.session.description",
								Other: "enter the session ID",
							}),
							AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
								ID:    "field.session.hint",
								Other: "Session ID",
							}),
							AutocompletePosition: 1,
							IsRequired:           true,
						},
					},
				},
			}
		},

		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			sessionID := creq.GetValue(fSessionID, "")
			session, err := a.conf.MattermostAPI().Session.Get(sessionID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewTextResponse(utils.JSONBlock(session))
		},
	}
}
