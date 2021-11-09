// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/session"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugSessionsList() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
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
				Call: &apps.Call{
					Path: pDebugSessionsList,
					Expand: &apps.Expand{
						ActingUser: apps.ExpandSummary,
					},
				},
				Form: &noParameters,
			}
		},

		submitf: func(_ context.Context, creq apps.CallRequest) apps.CallResponse {
			loc := a.newLocalizer(creq)
			sessions, err := a.sessionService.ListForUser(creq.Context.ActingUserID)
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
				appID := session.GetAppID(s)
				if appID == "" {
					// Assume it's the builtin app
					appID = AppID
				}

				expiresAt := time.UnixMilli(s.ExpiresAt).String()
				expiresIn := time.Until(time.UnixMilli(s.ExpiresAt)).String()
				token := utils.LastN(s.Token, 4)

				txt += fmt.Sprintf("|%s|%s|%s|%s|%s|\n",
					sessionID, appID, expiresAt, expiresIn, token)
			}

			return apps.NewTextResponse(txt)
		},
	}
}
