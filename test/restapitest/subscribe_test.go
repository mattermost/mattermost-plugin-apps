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

func subApp(t testing.TB) *goapp.App {
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
			Subject:   apps.Subject(subject),
			TeamID:    teamID,
			ChannelID: channelID,
			Call: *apps.NewCall("/echo").
				WithExpand(apps.Expand{
					App:                   apps.ExpandAll,
					ActingUser:            apps.ExpandAll,
					ActingUserAccessToken: apps.ExpandAll,
					Locale:                apps.ExpandAll,
					Channel:               apps.ExpandAll,
					ChannelMember:         apps.ExpandAll,
					Team:                  apps.ExpandAll,
					TeamMember:            apps.ExpandAll,
					Post:                  apps.ExpandAll,
					RootPost:              apps.ExpandAll,
					User:                  apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
				}),
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

func subCall(th *Helper, path string, asBot bool, subject apps.Subject, teamID, channelID string) apps.CallResponse { //nolint:golint,unparam
	return *th.HappyCall(subID, subCallRequest(path, asBot, subject, teamID, channelID))
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
	th.InstallApp(subApp(th.T))

	th.Run("Unauthenticated requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.CreateUnauthenticatedClientPP()

		resp, err := client.Subscribe(&apps.Subscription{Subject: apps.SubjectUserCreated})
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		resp, err = client.Unsubscribe(&apps.Subscription{Subject: apps.SubjectUserCreated})
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		subs, resp, err := client.GetSubscriptions()
		assert.Empty(subs)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("subscribe-list-delete", func(th *Helper) {
		for _, asBot := range []bool{true, false} {
			name := "as acting user"
			if asBot {
				name = "as bot"
			}
			th.Run(name, func(th *Helper) {
				require := require.New(th)

				cresp := subCall(th, "/subscribe", asBot, apps.SubjectUserCreated, "", "")
				require.Equal(`subscribed`, cresp.Text)

				cresp = subCall(th, "/list", asBot, "", "", "")
				subs := []apps.Subscription{}
				err := json.Unmarshal([]byte(cresp.Text), &subs)
				require.NoError(err)
				require.Equal(1, len(subs))
				require.NotEmpty(subs[0].UserID)
				require.Equal(apps.SubjectUserCreated, subs[0].Subject)
				require.Equal(subID, subs[0].AppID)

				cresp = subCall(th, "/unsubscribe", asBot, apps.SubjectUserCreated, "", "")
				require.Equal(`unsubscribed`, cresp.Text)

				cresp = subCall(th, "/list", asBot, "", "", "")
				subs = []apps.Subscription{}
				err = json.Unmarshal([]byte(cresp.Text), &subs)
				require.NoError(err)
				require.Equal(0, len(subs))
			})
		}
	})

	th.Run("user-namespace", func(th *Helper) {
		require := require.New(th)

		// subscribe both users to user_created and channel_created.
		creq := subCallRequest("/subscribe", false, apps.SubjectUserCreated, "", "")
		cresp := th.HappyCall(subID, creq)
		require.Equal(`subscribed`, cresp.Text)
		cresp = th.HappyUser2Call(subID, creq)
		require.Equal(`subscribed`, cresp.Text)

		creq = subCallRequest("/subscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, "")
		cresp = th.HappyCall(subID, creq)
		require.Equal(`subscribed`, cresp.Text)
		cresp = th.HappyUser2Call(subID, creq)
		require.Equal(`subscribed`, cresp.Text)

		creq = subCallRequest("/list", false, "", "", "")
		cresp = th.HappyCall(subID, creq)
		subs := []apps.Subscription{}
		err := json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(2, len(subs))
		cresp = th.HappyUser2Call(subID, creq)
		subs = []apps.Subscription{}
		err = json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(2, len(subs))

		// Unsubscribe user1 from user_created, check.
		creq = subCallRequest("/unsubscribe", false, apps.SubjectUserCreated, "", "")
		cresp = th.HappyCall(subID, creq)
		require.Equal(`unsubscribed`, cresp.Text)
		creq = subCallRequest("/list", false, "", "", "")
		cresp = th.HappyCall(subID, creq)
		subs = []apps.Subscription{}
		err = json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(1, len(subs))
		require.Equal(apps.SubjectChannelCreated, subs[0].Subject)
		cresp = th.HappyUser2Call(subID, creq)
		subs = []apps.Subscription{}
		err = json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(2, len(subs))

		// Unsubscribe user2 from channel_created, cross-check.
		creq = subCallRequest("/unsubscribe", false, apps.SubjectChannelCreated, th.ServerTestHelper.BasicTeam.Id, "")
		cresp = th.HappyUser2Call(subID, creq)
		require.Equal(`unsubscribed`, cresp.Text)
		creq = subCallRequest("/list", false, "", "", "")
		cresp = th.HappyUser2Call(subID, creq)
		subs = []apps.Subscription{}
		err = json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(1, len(subs))
		require.Equal(apps.SubjectUserCreated, subs[0].Subject)
		cresp = th.HappyCall(subID, creq)
		subs = []apps.Subscription{}
		err = json.Unmarshal([]byte(cresp.Text), &subs)
		require.NoError(err)
		require.Equal(1, len(subs))
		require.Equal(apps.SubjectChannelCreated, subs[0].Subject)
	})
}
