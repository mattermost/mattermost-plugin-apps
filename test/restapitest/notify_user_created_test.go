// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func testNotifyUserCreated(app *apps.App, appBotUser *model.User, received chan apps.CallRequest) func(*Helper) {
	return func(th *Helper) {
		require := require.New(th)
		// Make sure the bot is a team and a channel member to be able to
		// subscribe and be notified; the user already is, and sysadmin can see
		// everything.
		tm, resp, err := th.ServerTestHelper.Client.AddTeamMember(th.ServerTestHelper.BasicTeam.Id, app.BotUserID)
		require.NoError(err)
		require.Equal(th.ServerTestHelper.BasicTeam.Id, tm.TeamId)
		require.Equal(app.BotUserID, tm.UserId)
		api4.CheckCreatedStatus(th, resp)

		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(th.ServerTestHelper.BasicChannel.Id, app.BotUserID)
		require.NoError(err)
		require.Equal(th.ServerTestHelper.BasicChannel.Id, cm.ChannelId)
		require.Equal(app.BotUserID, cm.UserId)
		api4.CheckCreatedStatus(th, resp)

		for _, level := range []apps.ExpandLevel{apps.ExpandNone, apps.ExpandID, apps.ExpandSummary, apps.ExpandAll} {
			name := string(level)
			if name == "" {
				name = "none"
			}
			th.Run("expand "+name, func(th *Helper) {
				for _, cl := range allClientCombinations(th, appBotUser) {
					th.Run(cl.name, func(th *Helper) {
						th.subscribeAs(cl, app.AppID, apps.Event{Subject: apps.SubjectUserCreated}, expandEverything(level))
						user := th.createTestUser()
						n := <-received
						require.Empty(received)
						require.EqualValues(apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)
						require.Equal(apps.SubjectUserCreated, n.Context.Subject)

						ec := apps.ExpandedContext{}
						switch level {
						case apps.ExpandID:
							ec = apps.ExpandedContext{
								User: &model.User{
									Id: user.Id,
								},
								ActingUser: &model.User{
									Id: cl.expectedActingUser.Id,
								},
								Locale: "en",
							}

						case apps.ExpandSummary:
							ec = apps.ExpandedContext{
								App: &apps.App{
									Manifest: apps.Manifest{
										AppID:   app.AppID,
										Version: app.Version,
									},
									BotUserID:   app.BotUserID,
									BotUsername: app.BotUsername,
								},
								User: &model.User{
									Email:     user.Email,
									FirstName: user.FirstName,
									Id:        user.Id,
									LastName:  user.LastName,
									Locale:    user.Locale,
									Nickname:  user.Nickname,
									Roles:     user.Roles,
									Timezone:  user.Timezone,
									Username:  user.Username,
								},
								ActingUser: &model.User{
									BotDescription: cl.expectedActingUser.BotDescription,
									Email:          cl.expectedActingUser.Email,
									FirstName:      cl.expectedActingUser.FirstName,
									Id:             cl.expectedActingUser.Id,
									IsBot:          cl.expectedActingUser.IsBot,
									LastName:       cl.expectedActingUser.LastName,
									Locale:         cl.expectedActingUser.Locale,
									Nickname:       cl.expectedActingUser.Nickname,
									Roles:          cl.expectedActingUser.Roles,
									Timezone:       cl.expectedActingUser.Timezone,
									Username:       cl.expectedActingUser.Username,
								},
								Locale: "en",
							}

						case apps.ExpandAll:
							clone := *app
							ec = apps.ExpandedContext{
								App:        &clone,
								User:       user,
								ActingUser: cl.expectedActingUser,
								Locale:     "en",
							}

							// Non-admins don't get the app's webhook secret expanded.
							if !cl.appActsAsSystemAdmin {
								ec.App.WebhookSecret = ""
							}
						}

						th.verifyContext(level, app, cl.appActsAsSystemAdmin,
							apps.Context{
								Subject:         apps.SubjectUserCreated,
								ExpandedContext: ec,
							},
							n.Context)
					})
				}
			})
		}
	}
}
