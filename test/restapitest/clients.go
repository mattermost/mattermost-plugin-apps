// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"fmt"
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
		th.UserClientPP.GetPluginRoute(appclient.AppsPluginName)+appspath.API+appspath.Call, string(b)) // nolint:bodyclose
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
type clientCombination struct {
	name         string
	happyCall    func(appID apps.AppID, creq apps.CallRequest) *apps.CallResponse
	call         func(appID apps.AppID, creq apps.CallRequest) (*apps.CallResponse, *model.Response, error)
	appActsAsBot bool
}

func userAsBotClientCombination(th *Helper) clientCombination {
	return clientCombination{
		name:         "user as bot",
		happyCall:    th.HappyCall,
		call:         th.Call,
		appActsAsBot: true,
	}
}

func adminAsBotClientCombination(th *Helper) clientCombination {
	return clientCombination{
		name:         "admin as bot",
		happyCall:    th.HappyAdminCall,
		call:         th.AdminCall,
		appActsAsBot: true,
	}
}

func userClientCombination(th *Helper) clientCombination {
	return clientCombination{

		name:      "user",
		happyCall: th.HappyCall,
		call:      th.Call,
	}
}

func user2ClientCombination(th *Helper) clientCombination {
	return clientCombination{
		name:      "user2",
		happyCall: th.HappyUser2Call,
		call:      th.User2Call,
	}
}

func adminClientCombination(th *Helper) clientCombination {
	return clientCombination{
		name:      "admin",
		happyCall: th.HappyAdminCall,
		call:      th.AdminCall,
	}
}

func allClientCombinations(th *Helper) []clientCombination {
	return []clientCombination{
		userAsBotClientCombination(th),
		adminAsBotClientCombination(th),
		userClientCombination(th),
		user2ClientCombination(th),
		adminClientCombination(th),
	}
}
