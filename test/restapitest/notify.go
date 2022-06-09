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
)

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

func triggerUserCreated() func(*Helper) apps.ExpandedContext {
	return func(th *Helper) apps.ExpandedContext {
		user := th.createTestUser()
		return apps.ExpandedContext{
			User: user,
		}
	}
}

func verifyUserCreated() func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
	return func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		user, resp, err := th.ServerTestHelper.SystemAdminClient.GetUser(data.User.Id, "")
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)
		return apps.ExpandedContext{
			User: user,
		}
	}
}

func triggerBotJoinedChannel(botUserID string) func(*Helper) apps.ExpandedContext {
	return func(th *Helper) apps.ExpandedContext {
		ch := th.createTestChannel(th.ServerTestHelper.BasicTeam.Id)
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, botUserID)
		require.NoError(th, err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added app's bot to channel %s", ch.Name)
		return apps.ExpandedContext{
			Channel:       ch,
			ChannelMember: cm,
		}
	}
}

func verifyBotJoinedChannel(appBotUser *model.User) func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
	return func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		// Get updated values.
		cm, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannelMember(data.Channel.Id, appBotUser.Id, "")
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)
		channel, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannel(data.Channel.Id, "")
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)
		return apps.ExpandedContext{
			Channel:       channel,
			ChannelMember: cm,
			User:          appBotUser,
		}
	}
}

func triggerBotLeftChannel(botUserID string) func(*Helper) apps.ExpandedContext {
	return func(th *Helper) apps.ExpandedContext {
		ec := triggerBotJoinedChannel(botUserID)(th)
		resp, err := th.ServerTestHelper.Client.RemoveUserFromChannel(ec.ChannelMember.ChannelId, ec.ChannelMember.UserId)
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)
		th.Logf("removed app's bot from channel %s", ec.Channel.Name)
		return apps.ExpandedContext{
			Channel: ec.Channel,
		}
	}
}

const expectExpandedChannel = true
const expectNoExpandedChannel = false

func verifyBotLeftChannel(appBotUser *model.User, expectChannel bool) func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
	return func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
		if !expectChannel {
			return apps.ExpandedContext{
				User: appBotUser,
			}
		}

		// Get updated values.
		channel, resp, err := th.ServerTestHelper.SystemAdminClient.GetChannel(data.Channel.Id, "")
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)

		return apps.ExpandedContext{
			Channel: channel,
			User:    appBotUser,
		}
	}
}

//
//
//

func (th *Helper) triggerUserJoinedChannel(ch *model.Channel, user *model.User) *model.ChannelMember {
	_ = th.triggerUserJoinedTeam(ch.TeamId, user)
	cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, user.Id)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s to channel %s", user.Username, ch.Name)
	return cm
}

func (th *Helper) triggerUserLeftChannel(ch *model.Channel, user *model.User) *model.ChannelMember {
	cm := th.triggerUserJoinedChannel(ch, user)
	_, err := th.ServerTestHelper.SystemAdminClient.RemoveUserFromChannel(ch.Id, cm.UserId)
	require.NoError(th, err)
	th.Logf("removed user @%s from channel %s)", user.Username, ch.Name)
	return cm
}

func (th *Helper) triggerUserJoinedTeam(teamID string, user *model.User) *model.TeamMember {
	tm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(teamID, user.Id)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s to team %s)", user.Username, teamID)
	return tm
}

func (th *Helper) triggerUserLeftTeam(teamID string, user *model.User) *model.TeamMember {
	tm := th.triggerUserJoinedTeam(teamID, user)
	_, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(teamID, user.Id)
	require.NoError(th, err)
	th.Logf("removed user @%s from team %s)", user.Username, teamID)
	return tm
}

func (th *Helper) triggerBotJoinedTeam(team *model.Team, botUserID string) *model.TeamMember {
	cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, botUserID)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added app's bot to team %s", team.Name)
	return cm
}

func (th *Helper) triggerBotLeftTeam(team *model.Team, botUserID string) *model.TeamMember {
	tm := th.triggerBotJoinedTeam(team, botUserID)
	resp, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(team.Id, botUserID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("removed app's bot from team %s", team.Name)
	return tm
}

func triggerChannelCreated(teamID string) TestFunc {
	return func(th *Helper) {
		_ = th.createTestChannel(teamID)
	}
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

func (th *Helper) createTestChannel(teamID string) *model.Channel {
	testName := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	ch, resp, err := th.ServerTestHelper.Client.CreateChannel(&model.Channel{
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
