// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//go:build e2e
// +build e2e

package restapi

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests

var pluginID = "com.mattermost.apps"

type TestHelper struct {
	ServerTestHelper    *api4.TestHelper
	ClientPP            *appclient.ClientPP
	SystemAdminClientPP *appclient.ClientPP
	BotClientPP         *appclient.ClientPP
	LocalClientPP       *appclient.ClientPP
}

func (th *TestHelper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func Setup(t testing.TB) *TestHelper {
	th := &TestHelper{}

	serverTestHelper := api4.Setup(t)
	serverTestHelper.InitBasic()

	// enable bot creation by default
	serverTestHelper.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.EnableBotAccountCreation = true
		*cfg.ServiceSettings.SiteURL = "http://localhost:8065"
	})

	th.ServerTestHelper = serverTestHelper

	th.ClientPP = th.CreateClientPP()
	th.ClientPP.AuthToken = th.ServerTestHelper.Client.AuthToken
	th.ClientPP.AuthType = th.ServerTestHelper.Client.AuthType
	th.SystemAdminClientPP = th.CreateClientPP()
	th.SystemAdminClientPP.AuthToken = th.ServerTestHelper.SystemAdminClient.AuthToken
	th.SystemAdminClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType
	th.LocalClientPP = th.CreateLocalClient("TODO")

	bot := th.ServerTestHelper.CreateBotWithSystemAdminClient()
	_, _, appErr := th.ServerTestHelper.App.AddUserToTeam(th.ServerTestHelper.Context, th.ServerTestHelper.BasicTeam.Id, bot.UserId, "")
	require.Nil(t, appErr)

	rtoken, _, err := th.ServerTestHelper.SystemAdminClient.CreateUserAccessToken(bot.UserId, "test token")
	require.NoError(t, err)

	th.BotClientPP = th.CreateClientPP()
	th.BotClientPP.AuthToken = rtoken.Token
	th.BotClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType

	return th
}

// Sets up the PP for test
func SetupPP(th *TestHelper, t testing.TB) {
	pluginsEnvironment := th.ServerTestHelper.App.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		mlog.Debug("Missing plugin environment")
		return
	}

	bundle := os.Getenv("PLUGIN_BUNDLE")
	require.NotEmpty(t, bundle, "PLUGIN_BUNDLE is not set, please run `make test-e2e`")

	require.NotEmpty(t, os.Getenv("MM_SERVER_PATH"), "MM_SERVER_PATH is not set, please set it to the path of your mattermost-server clone")

	// Install the PP and enable it
	pluginBytes, err := os.ReadFile(bundle)
	require.NoError(t, err)
	require.NotNil(t, pluginBytes)

	manifest, appErr := th.ServerTestHelper.App.InstallPlugin(bytes.NewReader(pluginBytes), true)
	require.Nil(t, appErr)
	require.Equal(t, pluginID, manifest.Id)

	_, err = pluginsEnvironment.Available()
	if err != nil {
		mlog.Error("Unable to get available plugins", mlog.Err(err))
		return
	}

	appErr = th.ServerTestHelper.App.EnablePlugin(pluginID)
	require.Nil(t, appErr)
	require.True(t, th.ServerTestHelper.App.GetPluginsEnvironment().IsActive(pluginID))
}

func (th *TestHelper) CreateClientPP() *appclient.ClientPP {
	return appclient.NewAppsPluginAPIClient(fmt.Sprintf("http://localhost:%v", th.ServerTestHelper.App.Srv().ListenAddr.Port))
}

func (th *TestHelper) CreateLocalClient(socketPath string) *appclient.ClientPP {
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

func (th *TestHelper) TestForUser(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"UserClientPP", func(t *testing.T) {
		f(t, th.ClientPP)
	})
}

func (th *TestHelper) TestForSystemAdmin(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"SystemAdminClientPP", func(t *testing.T) {
		f(t, th.SystemAdminClientPP)
	})
}

func (th *TestHelper) TestForLocal(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"LocalClientPP", func(t *testing.T) {
		f(t, th.LocalClientPP)
	})
}

func (th *TestHelper) TestForUserAndSystemAdmin(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	th.TestForUser(t, f)
	th.TestForSystemAdmin(t, f)
}
