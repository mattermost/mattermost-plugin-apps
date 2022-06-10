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

	basicTeam := th.ServerTestHelper.BasicTeam

	for name, tc := range map[string]struct {
		init               func(*Helper) apps.ExpandedContext
		event              func(*Helper, apps.ExpandedContext) apps.Event
		trigger            func(*Helper, apps.ExpandedContext) apps.ExpandedContext
		expected           func(*Helper, apps.ExpandedContext) apps.ExpandedContext
		clientCombinations []clientCombination
		expandCombinations []apps.ExpandLevel
	}{
		"all users receive user_created": {
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectUserCreated,
				}
			},
			trigger: func(th *Helper, _ apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					User: th.createTestUser(),
				}
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					User: th.getUser(data.User.Id),
				}
			},
		},

		"creator receives channel_created with ChannelMember and TeamMember": {
			clientCombinations: []clientCombination{
				userClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectChannelCreated,
					TeamID:  basicTeam.Id,
				}
			},
			trigger: func(th *Helper, _ apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id),
				}
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id),
					ChannelMember: th.getChannelMember(data.Channel.Id, th.ServerTestHelper.BasicUser.Id),
					Team:          th.getTeam(data.Channel.TeamId),
					TeamMember:    th.getTeamMember(data.Channel.TeamId, th.ServerTestHelper.BasicUser.Id),
				}
			},
		},

		"admin receives channel_created": {
			clientCombinations: []clientCombination{
				adminClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectChannelCreated,
					TeamID:  basicTeam.Id,
				}
			},
			trigger: func(th *Helper, _ apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id),
				}
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.getChannel(data.Channel.Id),
					Team:    th.getTeam(data.Channel.TeamId),
				}
			},
		},

		"bot user and admin receive bot_joined_channel in basic team": {
			clientCombinations: []clientCombination{
				botClientCombination(th, appBotUser),
				adminClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotJoinedChannel,
					TeamID:  basicTeam.Id,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.createTestChannel(th.ServerTestHelper.SystemAdminClient, basicTeam.Id),
					User:    appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				data.ChannelMember = th.addChannelMember(data.Channel, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id),
					ChannelMember: th.getChannelMember(data.Channel.Id, data.User.Id),
					User:          th.getUser(data.User.Id),
				}
			},
		},

		"bot receives unexpanded bot_left_channel": {
			clientCombinations: []clientCombination{
				botClientCombination(th, appBotUser),
				userClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotLeftChannel,
					TeamID:  basicTeam.Id,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, basicTeam.Id)
				cm := th.addChannelMember(channel, appBotUser)
				return apps.ExpandedContext{
					Channel:       channel,
					ChannelMember: cm,
					User:          appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				th.removeUserFromChannel(data.Channel, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					User: th.getUser(data.User.Id),
				}
			},
		},

		"admin receives expanded bot_left_channel": {
			// the user and admin can still to expand the channel after having been removed.
			clientCombinations: []clientCombination{
				adminClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotLeftChannel,
					TeamID:  basicTeam.Id,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, basicTeam.Id)
				cm := th.addChannelMember(channel, appBotUser)
				return apps.ExpandedContext{
					Channel:       channel,
					ChannelMember: cm,
					User:          appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				th.removeUserFromChannel(data.Channel, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.getChannel(data.Channel.Id),
					User:    th.getUser(data.User.Id),
				}
			},
		},

		"bot admin receive bot_joined_team": {
			clientCombinations: []clientCombination{
				botClientCombination(th, appBotUser),
				adminClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotJoinedTeam,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				team := th.createTestTeam()
				return apps.ExpandedContext{
					Team: team,
					User: appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				data.TeamMember = th.addTeamMember(data.Team, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Team:       th.getTeam(data.Team.Id),
					TeamMember: th.getTeamMember(data.Team.Id, data.User.Id),
					User:       th.getUser(data.User.Id),
				}
			},
		},

		"bot user user2 receive unexpanded bot_left_team": {
			clientCombinations: []clientCombination{
				botClientCombination(th, appBotUser),
				userClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotLeftTeam,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				team := th.createTestTeam()
				tm := th.addTeamMember(team, appBotUser)
				return apps.ExpandedContext{
					Team:       team,
					TeamMember: tm,
					User:       appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				th.removeTeamMember(data.Team, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					User: th.getUser(data.User.Id),
				}
			},
		},

		"admin receives expanded bot_left_team": {
			clientCombinations: []clientCombination{
				adminClientCombination(th),
			},
			event: func(*Helper, apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject: apps.SubjectBotLeftTeam,
				}
			},
			init: func(th *Helper) apps.ExpandedContext {
				team := th.createTestTeam()
				tm := th.addTeamMember(team, appBotUser)
				return apps.ExpandedContext{
					Team:       team,
					TeamMember: tm,
					User:       appBotUser,
				}
			},
			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				th.removeTeamMember(data.Team, data.User)
				return data
			},
			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Team: th.getTeam(data.Team.Id),
					// we get a DeleteAt set (unlike the channel member that
					// fails after the member removal).
					TeamMember: th.getTeamMember(data.Team.Id, data.User.Id),
					User:       th.getUser(data.User.Id),
				}
			},
		},

		"admin and channel members receive user_joined_channel": {
			init: func(th *Helper) apps.ExpandedContext {
				// create it as User, add bot and User2 as members, Admin will
				// have access anyway.
				channel := th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id)
				th.addChannelMember(channel, appBotUser)
				th.addChannelMember(channel, th.ServerTestHelper.BasicUser2)
				testUser := th.createTestUser()
				th.addTeamMember(basicTeam, testUser)
				return apps.ExpandedContext{
					Channel: channel,
					User:    testUser,
				}
			},

			event: func(th *Helper, data apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject:   apps.SubjectUserJoinedChannel,
					ChannelID: data.Channel.Id,
				}
			},

			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				data.ChannelMember = th.addChannelMember(data.Channel, data.User)
				return data
			},

			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id),
					ChannelMember: th.getChannelMember(data.Channel.Id, data.User.Id),
					User:          th.getUser(data.User.Id),
				}
			},
		},

		"admin and channel members receive user_left_channel": { // others can not subscribe.
			clientCombinations: []clientCombination{
				userClientCombination(th),
				user2ClientCombination(th),
				adminClientCombination(th),
			},

			init: func(th *Helper) apps.ExpandedContext {
				// create it as User, add bot and User2 as members, Admin will
				// have access anyway.
				channel := th.createTestChannel(th.ServerTestHelper.Client, basicTeam.Id)
				th.addChannelMember(channel, th.ServerTestHelper.BasicUser2)
				testUser := th.createTestUser()
				th.addTeamMember(basicTeam, testUser)
				cm := th.addChannelMember(channel, testUser)
				return apps.ExpandedContext{
					Channel:       channel,
					ChannelMember: cm,
					User:          testUser,
				}
			},

			event: func(th *Helper, data apps.ExpandedContext) apps.Event {
				return apps.Event{
					Subject:   apps.SubjectUserLeftChannel,
					ChannelID: data.Channel.Id,
				}
			},

			trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				th.removeUserFromChannel(data.Channel, data.User)
				return data
			},

			expected: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
				return apps.ExpandedContext{
					Channel: th.getChannel(data.Channel.Id),
					User:    th.getUser(data.User.Id),
				}
			},
		},
	} {
		th.Run(name, func(th *Helper) {
			forExpandClientCombinations(th, appBotUser, tc.expandCombinations, tc.clientCombinations,
				func(th *Helper, level apps.ExpandLevel, cl clientCombination) {
					data := apps.ExpandedContext{}
					if tc.init != nil {
						data = tc.init(th)
					}

					event := tc.event(th, data)
					th.subscribeAs(cl, installedApp.AppID, event, expandEverything(level))

					data = tc.trigger(th, data)

					n := <-received
					require.Empty(th, received)
					require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

					expected := apps.Context{
						Subject:         event.Subject,
						ExpandedContext: tc.expected(th, data),
					}
					expected.ExpandedContext.App = installedApp
					expected.ExpandedContext.ActingUser = cl.expectedActingUser
					expected.ExpandedContext.Locale = "en"

					th.verifyContext(level, installedApp, cl.appActsAsSystemAdmin, expected, n.Context)
				})
		})
	}
}
