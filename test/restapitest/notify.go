// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

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

func (th *Helper) createTestUser() *model.User {
	testUsername := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	testEmail := fmt.Sprintf("%s@test.test", testUsername)
	u, resp, err := th.ServerTestHelper.SystemAdminClient.CreateUser(&model.User{
		Username: testUsername,
		Email:    testEmail,
	})
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test user @%s (%s)", u.Username, u.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.DeleteUser(u.Id)
		require.NoError(th, err)
		th.Logf("deleted test user @%s (%s)", u.Username, u.Id)
	})
	return u
}

func (th *Helper) createTestChannel(client *model.Client4, teamID string) *model.Channel {
	testName := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	ch, resp, err := client.CreateChannel(&model.Channel{
		Name:   testName,
		Type:   model.ChannelTypePrivate,
		TeamId: teamID,
	})
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test channel %s (%s)", ch.Name, ch.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.DeleteChannel(ch.Id)
		require.NoError(th, err)
		th.Logf("deleted test channel @%s (%s)", ch.Name, ch.Id)
	})
	return ch
}

func (th *Helper) createTestTeam() *model.Team {
	testName := fmt.Sprintf("test%v", rand.Int()) //nolint:gosec
	team, resp, err := th.ServerTestHelper.SystemAdminClient.CreateTeam(&model.Team{
		Name:        testName,
		DisplayName: testName,
		Type:        model.TeamOpen,
	})
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test team %s (%s)", team.Name, team.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.SoftDeleteTeam(team.Id)
		require.NoError(th, err)
		th.Logf("deleted test team @%s (%s)", team.Name, team.Id)
	})
	return team
}

func (th *Helper) addChannelMember(channel *model.Channel, user *model.User) *model.ChannelMember {
	cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddChannelMember(channel.Id, user.Id)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s (%s) to channel %s (%s)", user.Username, user.Id, channel.Name, channel.Id)
	return cm
}

func (th *Helper) addTeamMember(team *model.Team, user *model.User) *model.TeamMember {
	cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, user.Id)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s (%s) to team %s (%s)", user.Username, user.Id, team.Name, team.Id)
	return cm
}

func (th *Helper) removeUserFromChannel(channel *model.Channel, user *model.User) {
	resp, err := th.ServerTestHelper.SystemAdminClient.RemoveUserFromChannel(channel.Id, user.Id)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("removed user @%s (%s) from channel %s (%s)", user.Username, user.Id, channel.Name, channel.Id)
}

func (th *Helper) removeTeamMember(team *model.Team, user *model.User) {
	resp, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(team.Id, user.Id)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("removed user @%s (%s) from team %s (%s)", user.Username, user.Id, team.Name, team.Id)
}

func (th *Helper) getUser(userID string) *model.User {
	user, resp, err := th.ServerTestHelper.SystemAdminClient.GetUser(userID, "")
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	return user
}

func (th *Helper) getChannel(channelID string) *model.Channel {
	channel, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannel(channelID, "")
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	return channel
}

func (th *Helper) getChannelMember(channelID, userID string) *model.ChannelMember {
	cm, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannelMember(channelID, userID, "")
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	return cm
}

func (th *Helper) getTeam(channelID string) *model.Team {
	team, resp, err := th.ServerTestHelper.SystemAdminClient.GetTeam(channelID, "")
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	return team
}

func (th *Helper) getTeamMember(channelID, userID string) *model.TeamMember {
	tm, resp, err := th.ServerTestHelper.SystemAdminClient.GetTeamMember(channelID, userID, "")
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	return tm
}

func (th *Helper) subscribeAs(cl clientCombination, appID apps.AppID, event apps.Event, expand apps.Expand) {
	cresp := cl.happyCall(appID, apps.CallRequest{
		Call: *apps.NewCall("/subscribe").ExpandActingUserClient(),
		Values: map[string]interface{}{
			"sub": apps.Subscription{
				Event: event,
				Call:  *apps.NewCall("/notify").WithExpand(expand),
			},
			"as_bot": cl.appActsAsBot,
		},
	})
	require.Equal(th, `subscribed`, cresp.Text)
	th.Cleanup(func() {
		cresp := cl.happyCall(appID, apps.CallRequest{
			Call: *apps.NewCall("/unsubscribe").ExpandActingUserClient(),
			Values: map[string]interface{}{
				"sub": apps.Subscription{
					Event: event,
				},
				"as_bot": cl.appActsAsBot,
			},
		})
		require.Equal(th, `unsubscribed`, cresp.Text)
	})
}
