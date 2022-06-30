// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/sessionutils"
)

func (a *builtinApp) debugSessionsListBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "list",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.list.label",
			Other: "list",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.session.list.description",
			Other: "List all App specific sessions.",
		}),
		Submit: &apps.Call{
			Path: PathDebugSessionsList,
			Expand: &apps.Expand{
				ActingUser: apps.ExpandSummary,
			},
		},
	}
}

func (a *builtinApp) debugSessionsList(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	sessions, err := a.sessionService.ListForUser(r, creq.Context.ActingUser.Id)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	txt := a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.debug.session.list.submit.header",
		Other: "| SessionID | AppID | ExpiresAt | ExpiresIn | Token |",
	})
	txt += "\n| :--| :-- |:-- | :-- |\n"

	for _, s := range sessions {
		sessionID := s.Id
		appID := sessionutils.GetAppID(s)

		expiresAt := time.UnixMilli(s.ExpiresAt).String()
		expiresIn := time.Until(time.UnixMilli(s.ExpiresAt)).String()
		token := utils.LastN(s.Token, 4)

		txt += fmt.Sprintf("|%s|%s|%s|%s|%s|\n",
			sessionID, appID, expiresAt, expiresIn, token)
	}

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Text: txt,
		Data: sessions,
	}
}
