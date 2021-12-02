// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) debugSessionsDelete() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location: "delete",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.delete.label",
					Other: "delete",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.delete.description",
					Other: "Delete all App specific sessions.",
				}),
				Call: &apps.Call{
					Path: pDebugSessionsDelete,
					Expand: &apps.Expand{
						ActingUser: apps.ExpandSummary,
					},
				},
				Form: &noParameters,
			}
		},

		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)

			err := a.sessionService.DeleteAllForUser(r, r.ActingUserID())
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			return apps.NewTextResponse(a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.clean.submit",
				Other: "Deleted app specific sessions.",
			}))
		},
	}
}
