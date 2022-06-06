// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"math/rand"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
)

func (th *Helper) triggerUserJoinedChannel(ch *model.Channel, user *model.User) *model.ChannelMember {
	require := require.New(th)
	_ = th.triggerUserJoinedTeam(ch.TeamId, user)
	cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, user.Id)
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s to channel %s", user.Username, ch.Name)
	return cm
}

func (th *Helper) triggerUserLeftChannel(ch *model.Channel, user *model.User) *model.ChannelMember {
	require := require.New(th)
	cm := th.triggerUserJoinedChannel(ch, user)
	_, err := th.ServerTestHelper.SystemAdminClient.RemoveUserFromChannel(ch.Id, cm.UserId)
	require.NoError(err)
	th.Logf("removed user @%s from channel %s)", user.Username, ch.Name)
	return cm
}

func (th *Helper) triggerUserJoinedTeam(teamID string, user *model.User) *model.TeamMember {
	require := require.New(th)
	tm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(teamID, user.Id)
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s to team %s)", user.Username, teamID)
	return tm
}

func (th *Helper) triggerUserLeftTeam(teamID string, user *model.User) *model.TeamMember {
	require := require.New(th)
	tm := th.triggerUserJoinedTeam(teamID, user)
	_, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(teamID, user.Id)
	require.NoError(err)
	th.Logf("removed user @%s from team %s)", user.Username, teamID)
	return tm
}

func (th *Helper) triggerBotJoinedChannel(ch *model.Channel, botUserID string) *model.ChannelMember {
	require := require.New(th)
	cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, botUserID)
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added app's bot to channel %s", ch.Name)
	return cm
}

func (th *Helper) triggerBotLeftChannel(ch *model.Channel, botUserID string) *model.ChannelMember {
	require := require.New(th)
	cm := th.triggerBotJoinedChannel(ch, botUserID)
	resp, err := th.ServerTestHelper.Client.RemoveUserFromChannel(cm.ChannelId, cm.UserId)
	require.NoError(err)
	api4.CheckOKStatus(th, resp)
	th.Logf("removed app's bot from channel %s", ch.Name)
	return cm
}

func (th *Helper) triggerBotJoinedTeam(team *model.Team, botUserID string) *model.TeamMember {
	require := require.New(th)
	cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, botUserID)
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added app's bot to team %s", team.Name)
	return cm
}

func (th *Helper) triggerBotLeftTeam(team *model.Team, botUserID string) *model.TeamMember {
	require := require.New(th)
	tm := th.triggerBotJoinedTeam(team, botUserID)
	resp, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(team.Id, botUserID)
	require.NoError(err)
	api4.CheckOKStatus(th, resp)
	th.Logf("removed app's bot from team %s", team.Name)
	return tm
}

func (th *Helper) triggerChannelCreated(teamID string) *model.Channel {
	return th.createTestChannel(teamID)
}

func (th *Helper) createTestUser() *model.User {
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

func (th *Helper) createTestChannel(teamID string) *model.Channel {
	require := require.New(th)

	testName := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	ch, resp, err := th.ServerTestHelper.Client.CreateChannel(&model.Channel{
		Name:   testName,
		Type:   model.ChannelTypePrivate,
		TeamId: teamID,
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

func (th *Helper) createTestTeam() *model.Team {
	require := require.New(th)

	testName := fmt.Sprintf("test%v", rand.Int()) //nolint:gosec
	team, resp, err := th.ServerTestHelper.SystemAdminClient.CreateTeam(&model.Team{
		Name:        testName,
		DisplayName: testName,
		Type:        model.TeamOpen,
	})
	require.NoError(err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("created test team %s (%s)", team.Name, team.Id)
	th.Cleanup(func() {
		_, err := th.ServerTestHelper.SystemAdminClient.SoftDeleteTeam(team.Id)
		require.NoError(err)
		th.Logf("deleted test team @%s (%s)", team.Name, team.Id)
	})
	return team
}
