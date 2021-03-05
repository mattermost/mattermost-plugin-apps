// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
// +build e2e

package restapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests

var pluginID = "com.mattermost.apps"

type TestHelper struct {
	ServerTestHelper    *api4.TestHelper
	ClientPP            *mmclient.ClientPP
	SystemAdminClientPP *mmclient.ClientPP
	BotClientPP         *mmclient.ClientPP
	LocalClientPP       *mmclient.ClientPP
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
	})

	th.ServerTestHelper = serverTestHelper

	th.ClientPP = th.CreateClientPP()
	th.SystemAdminClientPP = th.CreateClientPP()
	th.SystemAdminClientPP.AuthToken = th.ServerTestHelper.SystemAdminClient.AuthToken
	th.SystemAdminClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType
	th.LocalClientPP = th.CreateLocalClient("TODO")

	bot := th.ServerTestHelper.CreateBotWithSystemAdminClient()
	_, err := th.ServerTestHelper.App.AddUserToTeam(th.ServerTestHelper.BasicTeam.Id, bot.UserId, "")
	require.Nil(t, err)

	rtoken, _ := th.ServerTestHelper.SystemAdminClient.CreateUserAccessToken(bot.UserId, "test token")

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

	// Install the PP and enable it
	pluginBytes, err := ioutil.ReadFile(bundle)
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

	_, _, activationErr := pluginsEnvironment.Activate(pluginID)
	require.NoError(t, activationErr)
	require.True(t, th.ServerTestHelper.App.GetPluginsEnvironment().IsActive(pluginID))
}

func (th *TestHelper) CreateClientPP() *mmclient.ClientPP {
	return mmclient.NewAPIClientPP(fmt.Sprintf("http://localhost:%v", th.ServerTestHelper.App.Srv().ListenAddr.Port))
}

func (th *TestHelper) CreateLocalClient(socketPath string) *mmclient.ClientPP {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return &mmclient.ClientPP{
		APIURL:     "http://_" + model.API_URL_SUFFIX,
		HTTPClient: httpClient,
	}
}

func (th *TestHelper) TestForSystemAdmin(t *testing.T, f func(*testing.T, *mmclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"SystemAdminClientPP", func(t *testing.T) {
		f(t, th.SystemAdminClientPP)
	})
}

func (th *TestHelper) TestForLocal(t *testing.T, f func(*testing.T, *mmclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"LocalClientPP", func(t *testing.T) {
		f(t, th.LocalClientPP)
	})
}
