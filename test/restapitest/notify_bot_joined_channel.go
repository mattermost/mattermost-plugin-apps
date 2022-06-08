// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func testNotifyBotJoinedChannel(app *apps.App, appBotUser *model.User, received chan apps.CallRequest) func(*Helper) {
	return func(th *Helper) {
		forExpandClientCombinations(th, appBotUser, nil,
			[]clientCombination{
				userAsBotClientCombination(th, appBotUser),
				adminAsBotClientCombination(th, appBotUser),
				userClientCombination(th),
				adminClientCombination(th),
			},
			func(th *Helper, level apps.ExpandLevel, cl clientCombination) {
				event := apps.Event{
					Subject: apps.SubjectBotJoinedChannel,
					TeamID:  th.ServerTestHelper.BasicTeam.Id,
				}
				th.subscribeAs(cl, app.AppID, event, expandEverything(level))

				ch := th.createTestChannel(th.ServerTestHelper.BasicTeam.Id)
				th.triggerBotJoinedChannel(ch, app.BotUserID)

				n := <-received
				require.Empty(th, received)
				require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

				// Get updated values.
				cm, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannelMember(ch.Id, appBotUser.Id, "")
				require.NoError(th, err)
				api4.CheckOKStatus(th, resp)

				channel, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannel(ch.Id, "")
				require.NoError(th, err)
				api4.CheckOKStatus(th, resp)

				th.verifyContext(level, app, cl.appActsAsSystemAdmin,
					apps.Context{
						Subject: apps.SubjectBotJoinedChannel,
						ExpandedContext: apps.ExpandedContext{
							ActingUser:    cl.expectedActingUser,
							App:           app,
							Channel:       channel,
							ChannelMember: cm,
							Locale:        "en",
							User:          appBotUser,
						},
					},
					n.Context)
			})
	}
}
