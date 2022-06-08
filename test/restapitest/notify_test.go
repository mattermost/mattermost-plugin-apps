// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func testNotify(th *Helper) {
	// 1000 is enough to receive all possible notifications that might come in a single test.
	received := make(chan apps.CallRequest, 1000)
	app := newNotifyApp(th, received)
	installedApp := th.InstallAppWithCleanup(app)
	rand.Seed(time.Now().UnixMilli())

	// Will need the bot user object later, preload.
	appBotUser, appErr := th.ServerTestHelper.App.GetUser(installedApp.BotUserID)
	require.Nil(th, appErr)

	// Make sure the bot is a team and a channel member to be able to
	// subscribe and be notified; the user already is, and sysadmin can see
	// everything.
	tm, resp, err := th.ServerTestHelper.Client.AddTeamMember(th.ServerTestHelper.BasicTeam.Id, installedApp.BotUserID)
	require.NoError(th, err)
	require.Equal(th, th.ServerTestHelper.BasicTeam.Id, tm.TeamId)
	require.Equal(th, installedApp.BotUserID, tm.UserId)
	api4.CheckCreatedStatus(th, resp)

	cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(th.ServerTestHelper.BasicChannel.Id, installedApp.BotUserID)
	require.NoError(th, err)
	require.Equal(th, th.ServerTestHelper.BasicChannel.Id, cm.ChannelId)
	require.Equal(th, installedApp.BotUserID, cm.UserId)
	api4.CheckCreatedStatus(th, resp)

	for name, testF := range map[string]func(*Helper){
		"user created":       testNotifyUserCreated(installedApp, appBotUser, received),
		"bot joined channel": testNotifyBotJoinedChannel(installedApp, appBotUser, received),
	} {
		th.Run(name, testF)
	}

	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectChannelCreated,
	// 				TeamID:  th.ServerTestHelper.BasicTeam.Id,
	// 			},
	// 			triggerf: triggerChannelCreated(th.ServerTestHelper.BasicTeam.Id),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject:   apps.SubjectUserJoinedChannel,
	// 				ChannelID: th.ServerTestHelper.BasicChannel.Id,
	// 			},
	// 			triggerf: triggerUserJoinedChannel(th.ServerTestHelper.BasicChannel),
	// 			// user2 is not a member of the channel, so must be excluded.
	// 			clientCombinations: []clientCombination{
	// 				userAsBotClientCombination(th),
	// 				adminAsBotClientCombination(th),
	// 				userClientCombination(th),
	// 				adminClientCombination(th),
	// 			},
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject:   apps.SubjectUserLeftChannel,
	// 				ChannelID: th.ServerTestHelper.BasicChannel.Id,
	// 			},
	// 			triggerf: triggerUserLeftChannel(th.ServerTestHelper.BasicChannel),
	// 			clientCombinations: []clientCombination{
	// 				userAsBotClientCombination(th),
	// 				adminAsBotClientCombination(th),
	// 				userClientCombination(th),
	// 				adminClientCombination(th),
	// 			},
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectUserJoinedTeam,
	// 				TeamID:  th.ServerTestHelper.BasicTeam.Id,
	// 			},
	// 			triggerf: triggerUserJoinedTeam(),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectUserLeftTeam,
	// 				TeamID:  th.ServerTestHelper.BasicTeam.Id,
	// 			},
	// 			triggerf: triggerUserLeftTeam(),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectBotJoinedChannel,
	// 				TeamID:  th.ServerTestHelper.BasicTeam.Id,
	// 			},
	// 			triggerf: triggerBotJoinedChannel(th.ServerTestHelper.BasicTeam.Id, installedApp.BotUserID),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectBotLeftChannel,
	// 				TeamID:  th.ServerTestHelper.BasicTeam.Id,
	// 			},
	// 			triggerf: triggerBotLeftChannel(th.ServerTestHelper.BasicTeam.Id, installedApp.BotUserID),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectBotJoinedTeam,
	// 			},
	// 			triggerf: triggerBotJoinedTeam(installedApp.BotUserID),
	// 		},
	// 		{
	// 			event: apps.Event{
	// 				Subject: apps.SubjectBotLeftTeam,
	// 			},
	// 			triggerf: triggerBotLeftTeam(installedApp.BotUserID),
	// 		},
	// 	} {
	// 		th.Run(string(tc.event.Subject), func(th *Helper) {
	// 			if tc.expand == nil {
	// 				tc.expand = map[apps.Expand]func(interface{}) apps.ExpandedContext{}
	// 			}
	// 			tc.expand[apps.Expand{}] = func(interface{}) apps.ExpandedContext { return apps.ExpandedContext{} }

	// 			combinations := tc.clientCombinations
	// 			if combinations == nil {
	// 				combinations = allClientCombinations(th)
	// 			}

	// 			for _, cl := range combinations {
	// 				th.Run(cl.name, func(th *Helper) {
	// 					for expand, expectedExpanded := range tc.expand {
	// 						th.Run("exp-"+expand.String(), func(th *Helper) {
	// 							// Subscribe to the event.
	// 							cresp := cl.happyCall(appID, apps.CallRequest{
	// 								Call: *apps.NewCall("/subscribe").ExpandActingUserClient(),
	// 								Values: map[string]interface{}{
	// 									"sub": apps.Subscription{
	// 										Event: tc.event,
	// 										Call:  *apps.NewCall("/notify").WithExpand(expand),
	// 									},
	// 									"as_bot": cl.appActsAsBot,
	// 								},
	// 							})
	// 							require.Equal(`subscribed`, cresp.Text)
	// 							th.Cleanup(func() {
	// 								cresp := cl.happyCall(appID, apps.CallRequest{
	// 									Call: *apps.NewCall("/unsubscribe").ExpandActingUserClient(),
	// 									Values: map[string]interface{}{
	// 										"sub": apps.Subscription{
	// 											Event: tc.event,
	// 										},
	// 										"as_bot": cl.appActsAsBot,
	// 									},
	// 								})
	// 								require.Equal(`unsubscribed`, cresp.Text)
	// 							})

	// 							// Make the event happen.
	// 							data := tc.triggerf(th)

	// 							// Verify notification
	// 							n := <-received
	// 							require.Empty(received)
	// 							require.EqualValues(apps.NewCall("/notify").WithExpand(expand), &n.Call)

	// 							cc := n.Context
	// 							cc.ExpandedContext = apps.ExpandedContext{}
	// 							require.EqualValues(apps.Context{Subject: tc.event.Subject}, cc)

	// 							th.requireEqualExpandedContext(installedApp, expectedExpanded(data), n.Context.ExpandedContext)
	// 						})
	// 					}
	// 				})
	// 			}
	// 		})
	// 	}
	// })
}
