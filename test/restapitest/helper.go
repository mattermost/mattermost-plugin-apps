// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests
var pluginID = "com.mattermost.apps"

type Helper struct {
	*testing.T
	ServerTestHelper *api4.TestHelper

	UserClientPP        *appclient.ClientPP
	User2ClientPP       *appclient.ClientPP
	SystemAdminClientPP *appclient.ClientPP
	LocalClientPP       *appclient.ClientPP
}

func NewHelper(t *testing.T, apps ...*goapp.App) *Helper {
	require := require.New(t)
	// Check environment
	require.NotEmpty(os.Getenv("MM_SERVER_PATH"),
		"MM_SERVER_PATH is not set, please set it to the path of your mattermost-server clone")

	// Unset SiteURL, just in case it's set
	err := os.Unsetenv("MM_SERVICESETTINGS_SITEURL")
	require.NoError(err)

	// Setup Mattermost server (helper)
	serverTestHelper := api4.Setup(t)
	serverTestHelper.InitBasic()
	port := serverTestHelper.Server.ListenAddr.Port
	serverTestHelper.App.UpdateConfig(func(cfg *model.Config) {
		// Need to create plugin and app bots.
		*cfg.ServiceSettings.EnableBotAccountCreation = true

		// Need to create and use OAuth2 apps.
		*cfg.ServiceSettings.EnableOAuthServiceProvider = true

		// Need to make requests to other local servers (apps).
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "127.0.0.1"

		// // Enable debug logging into file. -- DOESN'T WORK?
		// *cfg.LogSettings.EnableFile = true
		// *cfg.LogSettings.FileLevel = "DEBUG"
		// *cfg.LogSettings.FileJson = true
		// *cfg.LogSettings.EnableConsole = true
		// *cfg.LogSettings.ConsoleLevel = "DEBUG"
		// *cfg.LogSettings.ConsoleJson = true

		// Update the server own address, as we know it.
		*cfg.ServiceSettings.SiteURL = fmt.Sprintf("http://localhost:%d", port)
		*cfg.ServiceSettings.ListenAddress = fmt.Sprintf(":%d", port)
	})

	// TODO: <>/<> remove this later For the time being, there is a check
	// performed on the useragentcontext that prevents us from sending empty or
	// invalid IDs. So, add the users to the team/channel so they validate
	// correctly.
	initUser := func(id string) {
		_, resp, err := serverTestHelper.SystemAdminClient.AddTeamMember(serverTestHelper.BasicChannel.TeamId, id)
		require.NoError(err)
		api4.CheckCreatedStatus(t, resp)
		_, resp, err = serverTestHelper.SystemAdminClient.AddChannelMember(serverTestHelper.BasicChannel.Id, id)
		require.NoError(err)
		api4.CheckCreatedStatus(t, resp)
	}
	initUser(serverTestHelper.BasicUser.Id)
	initUser(serverTestHelper.BasicUser2.Id)
	initUser(serverTestHelper.SystemAdminUser.Id)

	// Create the helper and register for cleanup.
	th := &Helper{
		T:                t,
		ServerTestHelper: serverTestHelper,
	}
	t.Cleanup(th.TearDown)

	th.InitClients()
	th.InstallAppsPlugin()
	for _, a := range apps {
		th.InstallApp(a)
	}
	return th
}

func (th *Helper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func (th *Helper) InstallAppsPlugin() {
	require := require.New(th)

	bundle := os.Getenv("PLUGIN_BUNDLE")
	require.NotEmpty(bundle, "PLUGIN_BUNDLE is not set, please run `make test-rest-api`")

	// Install the PP and enable it
	pluginBytes, err := os.ReadFile(bundle)
	require.NoError(err)
	require.NotNil(pluginBytes)

	manifest, appErr := th.ServerTestHelper.App.InstallPlugin(bytes.NewReader(pluginBytes), true)
	require.Nil(appErr)
	require.Equal(pluginID, manifest.Id)

	appErr = th.ServerTestHelper.App.EnablePlugin(pluginID)
	require.Nil(appErr)
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

func (th *Helper) Run(name string, f func(th *Helper)) bool {
	return th.T.Run(name, func(t *testing.T) {
		h := *th
		h.T = t
		f(&h)
	})
}

func respond(text string, err error) apps.CallResponse {
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(text)
}
