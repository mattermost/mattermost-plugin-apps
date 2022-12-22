// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
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

type notifyTestCase struct {
	useTestSubscribe   bool
	init               func(*Helper, *model.User) apps.ExpandedContext
	event              func(*Helper, apps.ExpandedContext) apps.Event
	trigger            func(*Helper, apps.ExpandedContext) apps.ExpandedContext
	expected           func(*Helper, apps.ExpandLevel, appClient, apps.ExpandedContext) (apps.Subject, apps.ExpandedContext)
	except             []appClient
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

	params := func(creq goapp.CallRequest) (*appclient.Client, apps.Subscription, bool) {
		testFlag, _ := creq.BoolValue("test")
		asBot, _ := creq.BoolValue("as_bot")
		var sub apps.Subscription
		utils.Remarshal(&sub, creq.Values["sub"])

		if asBot {
			require.NotEmpty(th, creq.Context.BotAccessToken)
			return appclient.AsBot(creq.Context), sub, testFlag
		}
		require.NotEmpty(th, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), sub, testFlag
	}

	app.HandleCall("/subscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub, testFlag := params(creq)
			var err error
			if testFlag {
				err = client.SubscribeWithTestFlag(&sub)
			} else {
				err = client.Subscribe(&sub)
			}
			require.NoError(th, err)
			th.Logf("subscribed to %s", sub.Event)
			return apps.NewTextResponse("subscribed")
		})

	app.HandleCall("/unsubscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub, _ := params(creq)
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
		"bot_joined_channel_legacy":   notifyBotJoinedChannelLegacy(th),
		"bot_joined_channel_remapped": notifyBotJoinedChannelRemapped(th),
		"bot_joined_team_legacy":      notifyBotJoinedTeamLegacy(th),
		"bot_joined_team_remapped":    notifyBotJoinedTeamRemapped(th),
		"bot_left_channel_legacy":     notifyBotLeftChannelLegacy(th),
		"bot_left_channel_remapped":   notifyBotLeftChannelRemapped(th),
		"bot_left_team_legacy":        notifyBotLeftTeamLegacy(th),
		"bot_left_team_remapped":      notifyBotLeftTeamRemapped(th),
		"channel_created":             notifyChannelCreated(th),
		"self_joined_channel":         notifySelfJoinedChannel(th),
		"self_joined_team":            notifySelfJoinedTeam(th),
		"self_left_channel":           notifySelfLeftChannel(th),
		"self_left_team":              notifySelfLeftTeam(th),
		"user_created":                notifyUserCreated(th),
		"user_joined_channel":         notifyUserJoinedChannel(th),
		"user_joined_team":            notifyUserJoinedTeam(th),
		"user_left_channel":           notifyUserLeftChannel(th),
		"user_left_team":              notifyUserLeftTeam(th),
	} {
		th.Run(name, func(th *Helper) {
			forExpandClientCombinations(th, tc.expandCombinations, tc.except,
				func(th *Helper, level apps.ExpandLevel, appclient appClient) {
					data := apps.ExpandedContext{}
					if tc.init != nil {
						data = tc.init(th, appclient.expectedActingUser)
					}

					event := tc.event(th, data)
					th.subscribeAs(appclient, th.LastInstalledApp.AppID, event, expandEverything(level), tc.useTestSubscribe)

					data = tc.trigger(th, data)

					n := <-received
					require.Empty(th, received)
					require.EqualValues(th, apps.NewCall("/notify").WithExpand(expandEverything(level)), &n.Call)

					subj, ec := tc.expected(th, level, appclient, data)
					expected := apps.Context{
						Subject:         subj,
						ExpandedContext: ec,
					}
					expected.ExpandedContext.App = th.LastInstalledApp
					expected.ExpandedContext.ActingUser = appclient.expectedActingUser
					expected.ExpandedContext.Locale = "en"

					th.verifyContext(level, th.LastInstalledApp, appclient.appActsAsSystemAdmin, expected, n.Context)
				})
		})
	}
}

func (th *Helper) subscribeAs(appclient appClient, appID apps.AppID, event apps.Event, expand apps.Expand, testFlag bool) {
	cresp := appclient.happyCall(appID, apps.CallRequest{
		Call: *apps.NewCall("/subscribe").ExpandActingUserClient(),
		Values: map[string]interface{}{
			"sub": apps.Subscription{
				Event: event,
				Call:  *apps.NewCall("/notify").WithExpand(expand),
			},
			"as_bot": appclient.appActsAsBot,
			"test":   testFlag,
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
