// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const subID = apps.AppID("subtest")

func newSubscribeApp(t testing.TB) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       subID,
			Version:     "v1.1.0",
			DisplayName: "tests Subscription API",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
	)

	params := func(creq goapp.CallRequest) (*appclient.Client, apps.Subscription) {
		subject, _ := creq.StringValue("subject")
		channelID, _ := creq.StringValue("channel_id")
		teamID, _ := creq.StringValue("team_id")
		asBot, _ := creq.BoolValue("as_bot")

		sub := apps.Subscription{
			Event: apps.Event{
				Subject:   apps.Subject(subject),
				TeamID:    teamID,
				ChannelID: channelID,
			},
			Call: *apps.NewCall("/echo").
				WithExpand(expandEverything(apps.ExpandAll)),
		}

		if asBot {
			require.NotEmpty(t, creq.Context.BotAccessToken)
			return appclient.AsBot(creq.Context), sub
		}
		require.NotEmpty(t, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), sub
	}

	app.HandleCall("/subscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			return respond("subscribed", client.Subscribe(&sub))
		})

	app.HandleCall("/unsubscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			return respond("unsubscribed", client.Unsubscribe(&sub))
		})

	app.HandleCall("/list",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			subs, err := client.GetSubscriptions()
			return respond(utils.ToJSON(subs), err)
		})

	app.HandleCall("/echo", Echo)

	return app
}

func assertNumSubs(th *Helper, callf Caller, n int) {
	cresp := callf(subID, subCallRequest("/list", false, "", "", ""))
	subs := []apps.Subscription{}
	err := json.Unmarshal([]byte(cresp.Text), &subs)
	require.NoError(th, err)
	require.Equal(th, n, len(subs))
}

func subCallRequest(path string, asBot bool, subject apps.Subject, teamID, channelID string) apps.CallRequest {
	creq := apps.CallRequest{
		Call: *apps.NewCall(path),
		Values: model.StringInterface{
			"as_bot":    asBot,
			"subject":   subject,
			"team_id":   teamID,
			"channelID": channelID,
		},
	}
	if !asBot {
		creq.Call.Expand = &apps.Expand{
			ActingUser:            apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
		}
	}
	return creq
}

func testSubscriptions(th *Helper) {
	th.InstallAppWithCleanup(newSubscribeApp(th.T))

	th.Run("Unauthenticated requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.CreateUnauthenticatedClientPP()

		resp, err := client.Subscribe(&apps.Subscription{Event: apps.Event{Subject: apps.SubjectUserCreated}})
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		resp, err = client.Unsubscribe(&apps.Subscription{Event: apps.Event{Subject: apps.SubjectUserCreated}})
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		subs, resp, err := client.GetSubscriptions()
		assert.Empty(subs)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("subscribe-list-delete as user", func(th *Helper) {
		cresp := th.HappyCall(subID, subCallRequest("/subscribe", false, apps.SubjectUserCreated, "", ""))
		require.Equal(th, `subscribed`, cresp.Text)
		cresp = th.HappyCall(subID, subCallRequest("/subscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, ""))
		require.Equal(th, `subscribed`, cresp.Text)
		cresp = th.HappyCall(subID, subCallRequest("/subscribe", false, apps.SubjectUserJoinedTeam, th.ServerTestHelper.BasicTeam.Id, ""))
		require.Equal(th, `subscribed`, cresp.Text)
		assertNumSubs(th, th.HappyCall, 3)
		th.Cleanup(func() {
			_, _, _ = th.Call(subID, subCallRequest("/unsubscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, ""))
			_, _, _ = th.Call(subID, subCallRequest("/unsubscribe", false, apps.SubjectUserJoinedTeam, th.ServerTestHelper.BasicTeam.Id, ""))
			_, _, _ = th.Call(subID, subCallRequest("/unsubscribe", false, apps.SubjectUserCreated, "", ""))
			assertNumSubs(th, th.HappyCall, 0)
		})

		cresp = th.HappyCall(subID, subCallRequest("/unsubscribe", false, apps.SubjectUserCreated, "", ""))
		require.Equal(th, `unsubscribed`, cresp.Text)
		assertNumSubs(th, th.HappyCall, 2)

		// Unsubscribe from a non-existing subscription.
		badResponse, _, err := th.Call(subID, subCallRequest("/unsubscribe", false, apps.SubjectChannelCreated, "does-not-exist", ""))
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, badResponse.Type)
		require.Equal(th, "not found", badResponse.Text)
		assertNumSubs(th, th.HappyCall, 2)
	})

	th.Run("subscribe as bot list as user", func(th *Helper) {
		cresp := th.HappyCall(subID, subCallRequest("/subscribe", true, apps.SubjectUserCreated, "", ""))
		require.Equal(th, `subscribed`, cresp.Text)

		// listing subs as user, we don't see the bot subscription.
		assertNumSubs(th, th.HappyCall, 0)

		cresp = th.HappyCall(subID, subCallRequest("/unsubscribe", true, apps.SubjectUserCreated, "", ""))
		require.Equal(th, `unsubscribed`, cresp.Text)
	})

	th.Run("users have their namespaces", func(th *Helper) {
		// subscribe both users to user_created and channel_created.
		subscribeUserCreated := subCallRequest("/subscribe", false, apps.SubjectUserCreated, "", "")
		unsubscribeUserCreated := subCallRequest("/unsubscribe", false, apps.SubjectUserCreated, "", "")
		subscribeChannelCreated := subCallRequest("/subscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, "")
		unsubscribeChannelCreated := subCallRequest("/unsubscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, "")

		cresp := th.HappyCall(subID, subscribeUserCreated)
		require.Equal(th, `subscribed`, cresp.Text)
		cresp = th.HappyCall(subID, subscribeChannelCreated)
		require.Equal(th, `subscribed`, cresp.Text)
		assertNumSubs(th, th.HappyCall, 2)
		th.Cleanup(func() {
			_, _, _ = th.Call(subID, unsubscribeUserCreated)
			_, _, _ = th.Call(subID, unsubscribeChannelCreated)
			assertNumSubs(th, th.HappyCall, 0)
		})

		cresp = th.HappyUser2Call(subID, subscribeUserCreated)
		require.Equal(th, `subscribed`, cresp.Text)
		cresp = th.HappyUser2Call(subID, subscribeChannelCreated)
		require.Equal(th, `subscribed`, cresp.Text)
		assertNumSubs(th, th.HappyUser2Call, 2)
		th.Cleanup(func() {
			_, _, _ = th.User2Call(subID, unsubscribeUserCreated)
			_, _, _ = th.User2Call(subID, unsubscribeChannelCreated)
			assertNumSubs(th, th.HappyUser2Call, 0)
		})

		// Unsubscribe user1 from user_created, check.
		cresp = th.HappyCall(subID, unsubscribeUserCreated)
		require.Equal(th, `unsubscribed`, cresp.Text)
		assertNumSubs(th, th.HappyCall, 1)
		assertNumSubs(th, th.HappyUser2Call, 2)

		// Unsubscribe user2 from channel_created, cross-check.
		cresp = th.HappyUser2Call(subID, unsubscribeChannelCreated)
		require.Equal(th, `unsubscribed`, cresp.Text)
		assertNumSubs(th, th.HappyCall, 1)
		assertNumSubs(th, th.HappyUser2Call, 1)
	})
}
