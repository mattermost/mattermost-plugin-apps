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

type notifyTestCase struct {
	init               func(*Helper) apps.ExpandedContext
	event              func(*Helper, apps.ExpandedContext) apps.Event
	trigger            func(*Helper, apps.ExpandedContext) apps.ExpandedContext
	expected           func(*Helper, apps.ExpandLevel, appClient, apps.ExpandedContext) apps.ExpandedContext
	appClients         []appClient
	expandCombinations []apps.ExpandLevel
}

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
	rand.Seed(time.Now().UnixMilli())
	// 1000 is enough to receive all possible notifications that might come in a single test.
	received := make(chan apps.CallRequest, 1000)
	th.InstallAppWithCleanup(newNotifyApp(th, received))

	// Will need the bot user object later, preload.
	appBotUser, appErr := th.ServerTestHelper.App.GetUser(th.LastInstalledApp.BotUserID)
	require.Nil(th, appErr)
	th.LastInstalledBotUser = appBotUser

	// Make sure the bot is a team and a channel member to be able to
	// subscribe and be notified; the user already is, and sysadmin can see
	// everything.
	tm, resp, err := th.ServerTestHelper.Client.AddTeamMember(th.ServerTestHelper.BasicTeam.Id, th.LastInstalledApp.BotUserID)
	require.NoError(th, err)
	require.Equal(th, th.ServerTestHelper.BasicTeam.Id, tm.TeamId)
	require.Equal(th, th.LastInstalledApp.BotUserID, tm.UserId)
	api4.CheckCreatedStatus(th, resp)

	cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(th.ServerTestHelper.BasicChannel.Id, th.LastInstalledApp.BotUserID)
	require.NoError(th, err)
	require.Equal(th, th.ServerTestHelper.BasicChannel.Id, cm.ChannelId)
	require.Equal(th, th.LastInstalledApp.BotUserID, cm.UserId)
	api4.CheckCreatedStatus(th, resp)

	for name, tc := range map[string]*notifyTestCase{
		"bot_joined_channel": notifyBotJoinedChannel(th),
		"bot_left_channel":   notifyBotLeftChannel(th),
		"channel_created":    notifyChannelCreated(th),
		"user_created":       notifyUserCreated(th),

		// 	"bot receives unexpanded bot_left_channel": {
		// 		clientCombinations: []clientCombination{
		// 			botClientCombination(th, appBotUser),
		// 			userClientCombination(th),
		// 		},
		// 		event: func(*Helper, apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject: apps.SubjectBotLeftChannel,
		// 				TeamID:  basicTeam.Id,
		// 			}
		// 		},
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, basicTeam.Id)
		// 			cm := th.addChannelMember(channel, appBotUser)
		// 			return apps.ExpandedContext{
		// 				Channel:       channel,
		// 				ChannelMember: cm,
		// 				User:          appBotUser,
		// 			}
		// 		},
		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			th.removeUserFromChannel(data.Channel, data.User)
		// 			return data
		// 		},
		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				User: th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"admin receives expanded bot_left_channel": {
		// 		// the user and admin can still to expand the channel after having been removed.
		// 		clientCombinations: []clientCombination{
		// 			adminClientCombination(th),
		// 		},
		// 		event: func(*Helper, apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject: apps.SubjectBotLeftChannel,
		// 				TeamID:  basicTeam.Id,
		// 			}
		// 		},
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, basicTeam.Id)
		// 			cm := th.addChannelMember(channel, appBotUser)
		// 			return apps.ExpandedContext{
		// 				Channel:       channel,
		// 				ChannelMember: cm,
		// 				User:          appBotUser,
		// 			}
		// 		},
		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			th.removeUserFromChannel(data.Channel, data.User)
		// 			return data
		// 		},
		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				Channel:       th.getChannel(data.Channel.Id),
		// 				ChannelMember: data.ChannelMember, // ChannelMember is no longer available, so use the last we have.
		// 				User:          th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"bot admin receive bot_joined_team": {
		// 		clientCombinations: []clientCombination{
		// 			botClientCombination(th, appBotUser),
		// 			adminClientCombination(th),
		// 		},
		// 		event: func(*Helper, apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject: apps.SubjectBotJoinedTeam,
		// 			}
		// 		},
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			team := th.createTestTeam()
		// 			return apps.ExpandedContext{
		// 				Team: team,
		// 				User: appBotUser,
		// 			}
		// 		},
		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			data.TeamMember = th.addTeamMember(data.Team, data.User)
		// 			return data
		// 		},
		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				Team:       th.getTeam(data.Team.Id),
		// 				TeamMember: th.getTeamMember(data.Team.Id, data.User.Id),
		// 				User:       th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"bot user user2 receive unexpanded bot_left_team": {
		// 		clientCombinations: []clientCombination{
		// 			botClientCombination(th, appBotUser),
		// 			userClientCombination(th),
		// 		},
		// 		event: func(*Helper, apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject: apps.SubjectBotLeftTeam,
		// 			}
		// 		},
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			team := th.createTestTeam()
		// 			tm := th.addTeamMember(team, appBotUser)
		// 			return apps.ExpandedContext{
		// 				Team:       team,
		// 				TeamMember: tm,
		// 				User:       appBotUser,
		// 			}
		// 		},
		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			th.removeTeamMember(data.Team, data.User)
		// 			return data
		// 		},
		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				User: th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"admin receives expanded bot_left_team": {
		// 		clientCombinations: []clientCombination{
		// 			adminClientCombination(th),
		// 		},
		// 		event: func(*Helper, apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject: apps.SubjectBotLeftTeam,
		// 			}
		// 		},
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			team := th.createTestTeam()
		// 			tm := th.addTeamMember(team, appBotUser)
		// 			return apps.ExpandedContext{
		// 				Team:       team,
		// 				TeamMember: tm,
		// 				User:       appBotUser,
		// 			}
		// 		},
		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			th.removeTeamMember(data.Team, data.User)
		// 			return data
		// 		},
		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				Team: th.getTeam(data.Team.Id),
		// 				// we get a DeleteAt set (unlike the channel member that
		// 				// fails after the member removal).
		// 				TeamMember: th.getTeamMember(data.Team.Id, data.User.Id),
		// 				User:       th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"admin and channel members receive user_joined_channel": {
		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			// create it as User, add bot and User2 as members, Admin will
		// 			// have access anyway.
		// 			channel := th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id)
		// 			th.addChannelMember(channel, appBotUser)
		// 			th.addChannelMember(channel, th.ServerTestHelper.BasicUser2)
		// 			testUser := th.createTestUser()
		// 			th.addTeamMember(basicTeam, testUser)
		// 			return apps.ExpandedContext{
		// 				Channel: channel,
		// 				User:    testUser,
		// 			}
		// 		},

		// 		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject:   apps.SubjectUserJoinedChannel,
		// 				ChannelID: data.Channel.Id,
		// 			}
		// 		},

		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
		// 			return data
		// 		},

		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				Channel:       th.getChannel(data.Channel.Id),
		// 				ChannelMember: th.getChannelMember(data.Channel.Id, data.User.Id),
		// 				User:          th.getUser(data.User.Id),
		// 			}
		// 		},
		// 	},

		// 	"admin and channel members receive user_left_channel": { // others can not subscribe.
		// 		clientCombinations: []clientCombination{
		// 			userClientCombination(th),
		// 			user2ClientCombination(th),
		// 			adminClientCombination(th),
		// 		},

		// 		init: func(th *Helper) apps.ExpandedContext {
		// 			// create it as User, add bot and User2 as members, Admin will
		// 			// have access anyway.
		// 			channel := th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id)
		// 			th.addChannelMember(channel, th.ServerTestHelper.BasicUser2)
		// 			testUser := th.createTestUser()
		// 			th.addTeamMember(basicTeam, testUser)
		// 			cm := th.addChannelMember(channel, testUser)
		// 			return apps.ExpandedContext{
		// 				Channel:       channel,
		// 				ChannelMember: cm,
		// 				User:          testUser,
		// 			}
		// 		},

		// 		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
		// 			return apps.Event{
		// 				Subject:   apps.SubjectUserLeftChannel,
		// 				ChannelID: data.Channel.Id,
		// 			}
		// 		},

		// 		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			th.removeUserFromChannel(data.Channel, data.User)
		// 			return data
		// 		},

		// 		expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// 			return apps.ExpandedContext{
		// 				Channel: th.getChannel(data.Channel.Id),
		// 				User:    th.getUser(data.User.Id),
		// 			}
		// 		},
		// },
	} {
		th.Run(name, func(th *Helper) {
			forExpandClientCombinations(th, th.LastInstalledBotUser, tc.expandCombinations, tc.appClients,
				func(th *Helper, level apps.ExpandLevel, appclient appClient) {
					data := apps.ExpandedContext{}
					if tc.init != nil {
						data = tc.init(th)
					}

					event := tc.event(th, data)
					th.subscribeAs(appclient, th.LastInstalledApp.AppID, event, expandEverything(level))

					data = tc.trigger(th, data)

					n := <-received
					require.Empty(th, received)
					require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

					expected := apps.Context{
						Subject:         event.Subject,
						ExpandedContext: tc.expected(th, level, appclient, data),
					}
					expected.ExpandedContext.App = th.LastInstalledApp
					expected.ExpandedContext.ActingUser = appclient.expectedActingUser
					expected.ExpandedContext.Locale = "en"

					th.verifyContext(level, th.LastInstalledApp, appclient.appActsAsSystemAdmin, expected, n.Context)
				})
		})
	}
}

func (th *Helper) subscribeAs(appclient appClient, appID apps.AppID, event apps.Event, expand apps.Expand) {
	cresp := appclient.happyCall(appID, apps.CallRequest{
		Call: *apps.NewCall("/subscribe").ExpandActingUserClient(),
		Values: map[string]interface{}{
			"sub": apps.Subscription{
				Event: event,
				Call:  *apps.NewCall("/notify").WithExpand(expand),
			},
			"as_bot": appclient.appActsAsBot,
		},
	})
	require.Equal(th, `subscribed`, cresp.Text)
	th.Cleanup(func() {
		cresp := appclient.happyCall(appID, apps.CallRequest{
			Call: *apps.NewCall("/unsubscribe").ExpandActingUserClient(),
			Values: map[string]interface{}{
				"sub": apps.Subscription{
					Event: event,
				},
				"as_bot": appclient.appActsAsBot,
			},
		})
		require.Equal(th, `unsubscribed`, cresp.Text)
	})
}
