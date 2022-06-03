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

func triggerUserCreated() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		return createTestUserWithCleanup(th)
	}
}

func triggerUserJoinedChannel(ch *model.Channel) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		user := createTestUserWithCleanup(th)
		_ = addUserToBasicTeam(th, user)
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, user.Id)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added user @%s to channel %s", user.Username, ch.Name)

		return cm
	}
}

func triggerUserLeftChannel(ch *model.Channel) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		user := createTestUserWithCleanup(th)
		_ = addUserToBasicTeam(th, user)
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

func triggerUserJoinedTeam() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		return addUserToBasicTeam(th, createTestUserWithCleanup(th))
	}
}

func triggerUserLeftTeam() func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)
		user := createTestUserWithCleanup(th)
		_ = addUserToBasicTeam(th, user)
		_, err := th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(th.ServerTestHelper.BasicTeam.Id, user.Id)
		require.NoError(err)
		th.Logf("removed user @%s from team %s)", user.Username, th.ServerTestHelper.BasicTeam.Id)
		return nil
	}
}

func triggerBotJoinedChannel(teamID, botUserID string) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		ch := createTestChannel(th, teamID)
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, botUserID)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added app's bot to channel %s", ch.Name)
		return cm
	}
}

func triggerBotLeftChannel(teamID, botUserID string) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		ch := createTestChannel(th, teamID)
		cm, resp, err := th.ServerTestHelper.Client.AddChannelMember(ch.Id, botUserID)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added app's bot to channel %s", ch.Name)

		resp, err = th.ServerTestHelper.Client.RemoveUserFromChannel(ch.Id, botUserID)
		require.NoError(err)
		api4.CheckOKStatus(th, resp)
		th.Logf("removed app's bot from channel %s", ch.Name)

		return cm
	}
}

func triggerBotJoinedTeam(botUserID string) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		team := createTestTeam(th)
		cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, botUserID)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added app's bot to team %s", team.Name)

		return cm
	}
}

func triggerBotLeftTeam(botUserID string) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		require := require.New(th)

		team := createTestTeam(th)
		cm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, botUserID)
		require.NoError(err)
		api4.CheckCreatedStatus(th, resp)
		th.Logf("added app's bot to team %s", team.Name)

		resp, err = th.ServerTestHelper.SystemAdminClient.RemoveTeamMember(team.Id, botUserID)
		require.NoError(err)
		api4.CheckOKStatus(th, resp)
		th.Logf("removed app's bot from team %s", team.Name)

		return cm
	}
}

func triggerChannelCreated(teamID string) func(*Helper) interface{} {
	return func(th *Helper) interface{} {
		return createTestChannel(th, teamID)
	}
}

func createTestUserWithCleanup(th *Helper) *model.User {
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

func createTestChannel(th *Helper, teamID string) *model.Channel {
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

func createTestTeam(th *Helper) *model.Team {
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
