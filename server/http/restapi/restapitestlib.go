// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests

var pluginID = "com.mattermost.apps"

type TestHelper struct {
	ServerTestHelper    *api4.TestHelper
	SystemAdminClientPP *mmclient.ClientPP
	LocalClientPP       *mmclient.ClientPP
}

func (th *TestHelper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func Setup(tb testing.TB) *TestHelper {
	th := &TestHelper{}

	serverTestHelper := api4.Setup(tb)
	serverTestHelper.InitBasic()
	th.ServerTestHelper = serverTestHelper

	th.SystemAdminClientPP = th.CreateClientPP()
	th.SystemAdminClientPP.AuthToken = th.ServerTestHelper.SystemAdminClient.AuthToken
	th.SystemAdminClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType
	th.LocalClientPP = th.CreateLocalClient("TODO")
	return th
}

func SetupPP(t testing.TB) *TestHelper {
	th := Setup(t)

	pluginsEnvironment := th.ServerTestHelper.App.GetPluginsEnvironment()
	if pluginsEnvironment == nil {
		mlog.Debug("Missing plugin environment")
		return nil
	}

	if err := pluginsEnvironment.PerformHealthCheck(pluginID); err != nil {
		mlog.Debug("Missing proxy plugin")
		return nil
	}

	basePath := os.Getenv("MM_SERVER_PATH")
	testPluginPath := filepath.Join(basePath, model.PLUGIN_SETTINGS_DEFAULT_DIRECTORY, pluginID+".tar.gz")

	// Install the plugin and enable
	pluginBytes, err := ioutil.ReadFile(testPluginPath)
	require.NoError(t, err)
	require.NotNil(t, pluginBytes)

	manifest, appErr := th.ServerTestHelper.App.InstallPlugin(bytes.NewReader(pluginBytes), true)
	require.Nil(t, appErr)
	require.Equal(t, pluginID, manifest.Id)

	_, err = pluginsEnvironment.Available()
	if err != nil {
		mlog.Error("Unable to get available plugins", mlog.Err(err))
		return nil
	}

	_, _, activationErr := pluginsEnvironment.Activate(pluginID)
	require.NoError(t, activationErr)
	require.True(t, th.ServerTestHelper.App.GetPluginsEnvironment().IsActive(pluginID))

	return th
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
		ApiUrl:     "http://_" + model.API_URL_SUFFIX,
		HttpClient: httpClient,
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
