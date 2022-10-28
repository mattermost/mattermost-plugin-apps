// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type TestClientPP struct {
	*appclient.ClientPP
	UserID string
}

func (th *Helper) InitClients() {
	init := func(loginf func(*model.Client4)) *appclient.ClientPP {
		c := th.ServerTestHelper.CreateClient()
		loginf(c)

		cpp := th.CreateUnauthenticatedClientPP()
		cpp.AuthToken = c.AuthToken
		cpp.AuthType = c.AuthType
		return cpp
	}

	th.UserClientPP = init(th.ServerTestHelper.LoginBasicWithClient)
	th.User2ClientPP = init(th.ServerTestHelper.LoginBasic2WithClient)
	th.SystemAdminClientPP = init(th.ServerTestHelper.LoginSystemAdminWithClient)

	th.LocalClientPP = th.CreateLocalClient(*th.ServerTestHelper.App.Config().ServiceSettings.LocalModeSocketLocation)
}

func (th *Helper) CreateUnauthenticatedClientPP() *appclient.ClientPP {
	cfg := th.ServerTestHelper.App.Config()
	siteURL, err := url.Parse(*cfg.ServiceSettings.SiteURL)
	require.NoError(th, err)
	url := fmt.Sprintf("http://localhost:%v", th.ServerTestHelper.App.Srv().ListenAddr.Port) + siteURL.Path
	return appclient.NewAppsPluginAPIClient(url)
}

func (th *Helper) CreateLocalClient(socketPath string) *appclient.ClientPP {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	client := appclient.NewAppsPluginAPIClient("http://_" + model.APIURLSuffix)
	client.HTTPClient = httpClient

	return client
}

func (th *Helper) Call(appID apps.AppID, creq apps.CallRequest) (*apps.CallResponse, *model.Response, error) {
	creq.Context.UserAgentContext.AppID = appID
	creq.Context.UserAgentContext.UserAgent = "test"
	creq.Context.UserAgentContext.ChannelID = th.ServerTestHelper.BasicChannel.Id
	return th.UserClientPP.Call(creq)
}

func (th *Helper) CallWithAppMetadata(appID apps.AppID, creq apps.CallRequest) (*proxy.CallResponse, *model.Response, error) {
	creq.Context.UserAgentContext.AppID = appID
	creq.Context.UserAgentContext.UserAgent = "test"
	creq.Context.UserAgentContext.ChannelID = th.ServerTestHelper.BasicChannel.Id
	b, err := json.Marshal(&creq)
	if err != nil {
		return nil, nil, err
	}

	resp, err := th.UserClientPP.DoAPIPOST(
		th.UserClientPP.GetPluginRoute(appclient.AppsPluginName)+appspath.API+appspath.Call, string(b))
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, model.BuildResponse(resp), err
	}

	var cresp proxy.CallResponse
	err = json.NewDecoder(resp.Body).Decode(&cresp)
	if err != nil {
		return nil, model.BuildResponse(resp), errors.Wrap(err, "failed to decode response")
	}

	return &cresp, model.BuildResponse(resp), nil
}

func (th *Helper) User2Call(appID apps.AppID, creq apps.CallRequest) (*apps.CallResponse, *model.Response, error) {
	creq.Context.UserAgentContext.AppID = appID
	creq.Context.UserAgentContext.UserAgent = "test"
	creq.Context.UserAgentContext.ChannelID = th.ServerTestHelper.BasicChannel.Id
	return th.User2ClientPP.Call(creq)
}

func (th *Helper) AdminCall(appID apps.AppID, creq apps.CallRequest) (*apps.CallResponse, *model.Response, error) {
	creq.Context.UserAgentContext.AppID = appID
	creq.Context.UserAgentContext.UserAgent = "test"
	creq.Context.UserAgentContext.ChannelID = th.ServerTestHelper.BasicChannel.Id
	return th.SystemAdminClientPP.Call(creq)
}

func (th *Helper) HappyCall(appID apps.AppID, creq apps.CallRequest) *apps.CallResponse {
	cresp, resp, err := th.Call(appID, creq)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	require.True(th, cresp.Type != apps.CallResponseTypeError, "Error: %s", cresp.Text)
	return cresp
}

func (th *Helper) HappyUser2Call(appID apps.AppID, creq apps.CallRequest) *apps.CallResponse {
	cresp, resp, err := th.User2Call(appID, creq)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	require.True(th, cresp.Type != apps.CallResponseTypeError, "Error: %s", cresp.Text)
	return cresp
}

func (th *Helper) HappyAdminCall(appID apps.AppID, creq apps.CallRequest) *apps.CallResponse {
	cresp, resp, err := th.AdminCall(appID, creq)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	require.True(th, cresp.Type != apps.CallResponseTypeError, "Error: %s", cresp.Text)
	return cresp
}

// All calls to the App are made by invoking the /api/v1/call API, and must be
// made in the context of a user.
type appClient struct {
	name                 string
	expectedActingUser   *model.User
	happyCall            func(appID apps.AppID, creq apps.CallRequest) *apps.CallResponse
	call                 func(appID apps.AppID, creq apps.CallRequest) (*apps.CallResponse, *model.Response, error)
	appActsAsBot         bool
	appActsAsSystemAdmin bool
}

func (th *Helper) createTestUser() *model.User {
	testUsername := fmt.Sprintf("test_%v", rand.Int()) //nolint:gosec
	testEmail := fmt.Sprintf("%s@test.test", testUsername)
	u, resp, err := th.ServerTestHelper.SystemAdminClient.CreateUser(&model.User{
		Username: testUsername,
		Email:    testEmail,
		Password: "Pa$$word11",
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
	tm, resp, err := th.ServerTestHelper.SystemAdminClient.AddTeamMember(team.Id, user.Id)
	require.NoError(th, err)
	api4.CheckCreatedStatus(th, resp)
	th.Logf("added user @%s (%s) to team %s (%s)", user.Username, user.Id, team.Name, team.Id)
	return tm
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
