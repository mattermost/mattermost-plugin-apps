// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func newNotifyApp(th *Helper, received chan apps.CallRequest) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       "testnotify",
			Version:     "v1.1.0",
			DisplayName: "Tests notifications",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
	)

	params := func(creq goapp.CallRequest) (*appclient.Client, apps.Subscription) {
		asBot, _ := creq.BoolValue("as_bot")
		var sub apps.Subscription
		utils.Remarshal(&sub, creq.Values["sub"])

		if asBot {
			require.NotEmpty(th, creq.Context.BotAccessToken)
			return appclient.AsBot(creq.Context), sub
		}
		require.NotEmpty(th, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), sub
	}

	app.HandleCall("/subscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			err := client.Subscribe(&sub)
			require.NoError(th, err)
			th.Logf("subscribed to %s", sub.Event)
			return apps.NewTextResponse("subscribed")
		})

	app.HandleCall("/unsubscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			err := client.Unsubscribe(&sub)
			require.NoError(th, err)
			th.Logf("unsubscribed from %s", sub.Event)
			return apps.NewTextResponse("unsubscribed")
		})

	app.HandleCall("/notify",
		func(creq goapp.CallRequest) apps.CallResponse {
			received <- creq.CallRequest
			return respond("OK", nil)
		})

	return app
}

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

	for _, tc := range []struct {
		event              apps.Event
		triggerF           func(*Helper) apps.ExpandedContext
		expectedF          func(*Helper, apps.ExpandedContext) apps.ExpandedContext
		clientCombinations []clientCombination
		expandCombinations []apps.ExpandLevel
	}{
		{
			event: apps.Event{
				Subject: apps.SubjectUserCreated,
			},
			triggerF:  triggerUserCreated(),
			expectedF: verifyUserCreated(),
		},

		{
			event: apps.Event{
				Subject: apps.SubjectBotJoinedChannel,
				TeamID:  th.ServerTestHelper.BasicTeam.Id,
			},
			clientCombinations: []clientCombination{
				userAsBotClientCombination(th, appBotUser),
				adminAsBotClientCombination(th, appBotUser),
				userClientCombination(th),
				adminClientCombination(th),
			},
			triggerF:  triggerBotJoinedChannel(appBotUser.Id),
			expectedF: verifyBotJoinedChannel(appBotUser),
		},

		{
			event: apps.Event{
				Subject: apps.SubjectBotLeftChannel,
				TeamID:  th.ServerTestHelper.BasicTeam.Id,
			},
			clientCombinations: []clientCombination{
				userAsBotClientCombination(th, appBotUser),
				adminAsBotClientCombination(th, appBotUser),
			},
			triggerF:  triggerBotLeftChannel(appBotUser.Id),
			expectedF: verifyBotLeftChannel(appBotUser, expectNoExpandedChannel), // the bot won't be able to expand the channel after having been removed.
		},
		{
			event: apps.Event{
				Subject: apps.SubjectBotLeftChannel,
				TeamID:  th.ServerTestHelper.BasicTeam.Id,
			},
			clientCombinations: []clientCombination{
				userClientCombination(th),
				adminClientCombination(th),
			},
			triggerF:  triggerBotLeftChannel(appBotUser.Id),
			expectedF: verifyBotLeftChannel(appBotUser, expectExpandedChannel),
		},
	} {
		th.Run(string(tc.event.Subject), func(th *Helper) {
			forExpandClientCombinations(th, appBotUser, tc.expandCombinations, tc.clientCombinations,
				func(th *Helper, level apps.ExpandLevel, cl clientCombination) {
					th.subscribeAs(cl, installedApp.AppID, tc.event, expandEverything(level))

					data := tc.triggerF(th)

					n := <-received
					require.Empty(th, received)
					require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

					expected := apps.Context{
						Subject:         tc.event.Subject,
						ExpandedContext: tc.expectedF(th, data),
					}
					expected.ExpandedContext.App = installedApp
					expected.ExpandedContext.ActingUser = cl.expectedActingUser
					expected.ExpandedContext.Locale = "en"

					th.verifyContext(level, installedApp, cl.appActsAsSystemAdmin, expected, n.Context)
				})
		})
	}
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
