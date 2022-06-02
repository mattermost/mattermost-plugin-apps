// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed"
	"fmt"
	"math/rand"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var expandAll = apps.Expand{
	App:                   apps.ExpandAll.Optional(),
	ActingUser:            apps.ExpandAll.Optional(),
	ActingUserAccessToken: apps.ExpandAll.Optional(),
	Locale:                apps.ExpandAll.Optional(),
	Channel:               apps.ExpandAll.Optional(),
	ChannelMember:         apps.ExpandAll.Optional(),
	Team:                  apps.ExpandAll.Optional(),
	TeamMember:            apps.ExpandAll.Optional(),
	Post:                  apps.ExpandAll.Optional(),
	RootPost:              apps.ExpandAll.Optional(),
	User:                  apps.ExpandAll.Optional(),
	OAuth2App:             apps.ExpandAll.Optional(),
	OAuth2User:            apps.ExpandAll.Optional(),
}

// var expandAllSummary = apps.Expand{
// 	App:                   apps.ExpandSummary.Optional(),
// 	ActingUser:            apps.ExpandSummary.Optional(),
// 	ActingUserAccessToken: apps.ExpandAll.Optional(),
// 	Locale:                apps.ExpandSummary.Optional(),
// 	Channel:               apps.ExpandSummary.Optional(),
// 	ChannelMember:         apps.ExpandSummary.Optional(),
// 	Team:                  apps.ExpandSummary.Optional(),
// 	TeamMember:            apps.ExpandSummary.Optional(),
// 	Post:                  apps.ExpandSummary.Optional(),
// 	RootPost:              apps.ExpandSummary.Optional(),
// 	User:                  apps.ExpandSummary.Optional(),
// 	OAuth2App:             apps.ExpandSummary.Optional(),
// 	OAuth2User:            apps.ExpandAll.Optional(),
// }

// var expandAllID = apps.Expand{
// 	App:                   apps.ExpandID.Optional(),
// 	ActingUser:            apps.ExpandID.Optional(),
// 	ActingUserAccessToken: apps.ExpandAll.Optional(),
// 	Locale:                apps.ExpandID.Optional(),
// 	Channel:               apps.ExpandID.Optional(),
// 	ChannelMember:         apps.ExpandID.Optional(),
// 	Team:                  apps.ExpandID.Optional(),
// 	TeamMember:            apps.ExpandID.Optional(),
// 	Post:                  apps.ExpandID.Optional(),
// 	RootPost:              apps.ExpandID.Optional(),
// 	User:                  apps.ExpandID.Optional(),
// 	OAuth2App:             apps.ExpandID.Optional(),
// 	OAuth2User:            apps.ExpandAll.Optional(),
// }

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

func createTestUser(th *Helper) *model.User {
	require := require.New(th)
	testUsername := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	testEmail := fmt.Sprintf("%s@test.test", testUsername)
	u, resp, err := th.ServerTestHelper.SystemAdminClient.CreateUser(&model.User{
		Username: testUsername,
		Email:    testEmail,
	})
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test user @%s (%s)", u.Username, u.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.DeleteUser(u.Id)
		require.NoError(err)
		th.Logf("deleted test user @%s (%s)", u.Username, u.Id)
	})
	return u
}

func addUserToBasicTeam(th *Helper, user *model.User) *model.TeamMember {
	require := require.New(th)
	teamID := th.ServerTestHelper.BasicTeam.Id
	tm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(teamID, user.Id)
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s to team %s)", user.Username, teamID)
	return tm
}

func createTestChannel(th *Helper) *model.Channel {
	require := require.New(th)

	testName := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	ch, resp, err := th.ServerTestHelper.Client.CreateChannel(&model.Channel{
		Name:   testName,
		Type:   model.ChannelTypePrivate,
		TeamId: th.ServerTestHelper.BasicTeam.Id,
	})
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test channel %s (%s)", ch.Name, ch.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.DeleteChannel(ch.Id)
		require.NoError(err)
		th.Logf("deleted test channel @%s (%s)", ch.Name, ch.Id)
	})
	return ch
}

func triggerUserCreated() func(*Helper) interface{} {
	return func(th *Helper) interface{} { return createTestUser(th) }
}

func triggerChannelCreated() func(*Helper) interface{} {
	return func(th *Helper) interface{} { return createTestChannel(th) }
}

func triggerUserJoinedChannel() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		user := createTestUser(th)
		_ = addUserToBasicTeam(th, user)
		ch := th.ServerTestHelper.BasicChannel
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, user.Id)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added user @%s to channel %s", user.Username, ch.Name)
		th.Cleanup(func() {
			_, err := th.ServerTestHelper.SystemAdminClient.RemoveUserFromChannel(ch.Id, user.Id)
			require.NoError(err)
			th.Logf("removed user @%s from channel %s)", user.Username, ch.Name)
		})

		return cm
	}
}

func triggerUserJoinedTeam() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		return addUserToBasicTeam(th, createTestUser(th))
	}
}

func triggerUserLeftTeam() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		user := createTestUser(th)
		_ = addUserToBasicTeam(th, user)
		_, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(th.ServerTestHelper.BasicTeam.Id, user.Id)
		require.NoError(err)
		th.Logf("removed user @%s from team %s)", user.Username, th.ServerTestHelper.BasicTeam.Id)
		return nil
	}
}

func triggerUserLeftChannel() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		user := createTestUser(th)
		_ = addUserToBasicTeam(th, user)
		ch := th.ServerTestHelper.BasicChannel
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, user.Id)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added user @%s to channel %s", user.Username, ch.Name)
		_, err = th.ServerTestHelper.SystemAdminClient.RemoveUserFromChannel(ch.Id, user.Id)
		require.NoError(err)
		th.Logf("removed user @%s from channel %s)", user.Username, ch.Name)
		return cm
	}
}

func testNotify(th *Helper) {
	th.Run("happy no expand", func(th *Helper) {
		// 1000 is enough to receive all possible notifications that might come in a single test.
		received := make(chan apps.CallRequest, 1000)
		app := newNotifyApp(th, received)
		installedApp := th.InstallAppWithCleanup(app)
		appID := app.Manifest.AppID
		require := require.New(th)
		rand.Seed(time.Now().UnixMilli())

		// Make sure the bot is a team member to be able to subscribe and be
		// notified; the user already is, and sysadmin can see everything.
		tm, resp, err := th.ServerTestHelper.Client.AddTeamMember(th.ServerTestHelper.BasicTeam.Id, installedApp.BotUserID)
		require.NoError(err)
		require.NotNil(th.ServerTestHelper.BasicTeam.Id, tm.TeamId)
		require.NotNil(installedApp.BotUserID, tm.UserId)
		api4.CheckCreatedStatus(th, resp)

		type TC struct {
			event              apps.Event
			triggerf           func(*Helper) interface{}
			clientCombinations []clientCombination
		}

		for _, tc := range []TC{
			{
				event:    apps.Event{Subject: apps.SubjectUserCreated},
				triggerf: triggerUserCreated(),
			},
			{
				event:    apps.Event{Subject: apps.SubjectChannelCreated, TeamID: th.ServerTestHelper.BasicTeam.Id},
				triggerf: triggerChannelCreated(),
			},
			{
				event:    apps.Event{Subject: apps.SubjectUserJoinedChannel, ChannelID: th.ServerTestHelper.BasicChannel.Id},
				triggerf: triggerUserJoinedChannel(),
				clientCombinations: []clientCombination{
					userClientCombination(th),
					adminClientCombination(th),
				},
			},
			{
				event:    apps.Event{Subject: apps.SubjectUserLeftChannel, ChannelID: th.ServerTestHelper.BasicChannel.Id},
				triggerf: triggerUserLeftChannel(),
				clientCombinations: []clientCombination{
					userClientCombination(th),
					adminClientCombination(th),
				},
			},
			{
				event:    apps.Event{Subject: apps.SubjectUserJoinedTeam, TeamID: th.ServerTestHelper.BasicTeam.Id},
				triggerf: triggerUserJoinedTeam(),
			},
			{
				event:    apps.Event{Subject: apps.SubjectUserLeftTeam, TeamID: th.ServerTestHelper.BasicTeam.Id},
				triggerf: triggerUserLeftTeam(),
			},
		} {
			th.Run(string(tc.event.Subject), func(th *Helper) {
				combinations := tc.clientCombinations
				if combinations == nil {
					combinations = allClientCombinations(th)
				}
				for _, cl := range combinations {
					th.Run(cl.name, func(th *Helper) {
						// Subscribe to the event.
						cresp := cl.happyCall(appID, apps.CallRequest{
							Call: *apps.NewCall("/subscribe").ExpandActingUserClient(),
							Values: map[string]interface{}{
								"sub": apps.Subscription{
									Event: tc.event,
									Call:  *apps.NewCall("/notify"),
								},
								"as_bot": cl.appActsAsBot,
							},
						})
						require.Equal(`subscribed`, cresp.Text)
						th.Cleanup(func() {
							cresp := cl.happyCall(appID, apps.CallRequest{
								Call: *apps.NewCall("/unsubscribe").ExpandActingUserClient(),
								Values: map[string]interface{}{
									"sub": apps.Subscription{
										Event: tc.event,
									},
									"as_bot": cl.appActsAsBot,
								},
							})
							require.Equal(`unsubscribed`, cresp.Text)
						})

						// Make the event happen.
						tc.triggerf(th)

						// Verify notification
						n := <-received
						require.Empty(received)
						require.EqualValues(*apps.NewCall("/notify"), n.Call)

						cc := n.Context
						cc.ExpandedContext = apps.ExpandedContext{}
						require.EqualValues(apps.Context{Subject: tc.event.Subject}, cc)

						ec := n.Context.ExpandedContext
						require.NotEmpty(ec.MattermostSiteURL)
						require.NotEmpty(ec.AppPath)
						require.NotEmpty(ec.BotUserID)
						require.NotEmpty(ec.BotAccessToken)
						ec.MattermostSiteURL = ""
						ec.BotAccessToken = ""
						ec.BotUserID = ""
						ec.AppPath = ""
						require.Empty(ec)
					})
				}
			})
		}
	})
}
