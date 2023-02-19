// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) debugSessionsRevokeBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "revoke",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.revoke.label",
			Other: "revoke",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.revoke.description",
			Other: "Revoke all App specific sessions.",
		}),
		Submit: &apps.Call{
			Path: pDebugSessionsRevoke,
			Expand: &apps.Expand{
				ActingUser: apps.ExpandSummary,
			},
		},
	}
}

func (a *builtinApp) debugSessionsRevoke(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)

	err := a.sessionService.RevokeSessionsForUser(r, r.ActingUserID())
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewTextResponse(a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.debug.session.revoke.submit",
		Other: "Revoked all of your app specific sessions.",
	}))
}
