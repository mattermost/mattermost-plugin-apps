// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
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
