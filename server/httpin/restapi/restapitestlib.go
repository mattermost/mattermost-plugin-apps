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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests

var pluginID = "com.mattermost.apps"

type TestHelper struct {
	t                   *testing.T
	ServerTestHelper    *api4.TestHelper
	UserClientPP        *appclient.ClientPP
	User2ClientPP       *appclient.ClientPP
	SystemAdminClientPP *appclient.ClientPP
	BotClientPP         *appclient.ClientPP
	LocalClientPP       *appclient.ClientPP
}

type TestApp struct {
	Manifest apps.Manifest
	AsUser   *appclient.ClientPP
	AsUser2  *appclient.ClientPP
	AsBot    *appclient.ClientPP
}

func (th *TestHelper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func Setup(t *testing.T) *TestHelper {
	// Unset SiteURL, just in case it's set
	err := os.Unsetenv("MM_SERVICESETTINGS_SITEURL")
	require.NoError(t, err)

	// Enable Apps Feature Flags via env variables as it can't be set via the config
	os.Setenv("MM_FEATUREFLAGS_APPSENABLED", "true")
	t.Cleanup(func() { _ = os.Unsetenv("MM_FEATUREFLAGS_APPSENABLED") })

	th := &TestHelper{
		t: t,
	}

	serverTestHelper := api4.Setup(t)
	serverTestHelper.InitBasic()

	t.Cleanup(th.TearDown)

	port := serverTestHelper.Server.ListenAddr.Port

	// enable bot creation by default
	serverTestHelper.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.ServiceSettings.EnableBotAccountCreation = true
		*cfg.ServiceSettings.EnableOAuthServiceProvider = true
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "127.0.0.1"
		*cfg.ServiceSettings.SiteURL = fmt.Sprintf("http://localhost:%d", port)
		*cfg.ServiceSettings.ListenAddress = fmt.Sprintf(":%d", port)
	})

	th.ServerTestHelper = serverTestHelper

	// TODO(Ben): Cleanup client creation
	th.UserClientPP = th.CreateClientPP()
	th.UserClientPP.AuthToken = th.ServerTestHelper.Client.AuthToken
	th.UserClientPP.AuthType = th.ServerTestHelper.Client.AuthType

	user2Client4 := th.ServerTestHelper.CreateClient()
	th.ServerTestHelper.LoginBasic2WithClient(user2Client4)
	th.User2ClientPP = th.CreateClientPP()
	th.User2ClientPP.AuthToken = user2Client4.AuthToken
	th.User2ClientPP.AuthType = user2Client4.AuthType

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
func (th *TestHelper) SetupPP() {
	require := require.New(th.t)

	bundle := os.Getenv("PLUGIN_BUNDLE")
	require.NotEmpty(bundle, "PLUGIN_BUNDLE is not set, please run `make test-e2e`")

	require.NotEmpty(os.Getenv("MM_SERVER_PATH"), "MM_SERVER_PATH is not set, please set it to the path of your mattermost-server clone")

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

// Sets up the PP for test
func (th *TestHelper) SetupApp(m apps.Manifest) TestApp {
	th.t.Helper()
	require := require.New(th.t)
	assert := assert.New(th.t)

	var (
		asUser  *appclient.Client
		asUser2 *appclient.Client
		asBot   *appclient.Client
	)

	router := mux.NewRouter()
	router.HandleFunc(apps.DefaultPing.Path, httputils.HandleJSON(apps.NewDataResponse(nil)))
	router.HandleFunc("/setup/user", func(w http.ResponseWriter, r *http.Request) {
		creq, err := apps.CallRequestFromJSONReader(r.Body)
		require.NoError(err)
		require.NotNil(creq)

		asUser = appclient.AsActingUser(creq.Context)
		asBot = appclient.AsBot(creq.Context)

		httputils.WriteJSON(w, apps.NewDataResponse(nil))
	})
	router.HandleFunc("/setup/user2", func(w http.ResponseWriter, r *http.Request) {
		creq, err := apps.CallRequestFromJSONReader(r.Body)
		require.NoError(err)
		require.NotNil(creq)

		asUser2 = appclient.AsActingUser(creq.Context)

		httputils.WriteJSON(w, apps.NewDataResponse(nil))
	})
	appServer := httptest.NewServer(router)
	th.t.Cleanup(appServer.Close)

	m.HTTP = &apps.HTTP{
		RootURL: appServer.URL,
		UseJWT:  false,
	}
	m.HomepageURL = appServer.URL
	m.RequestedPermissions = apps.Permissions{
		apps.PermissionActAsBot,
		apps.PermissionActAsUser,
	}

	err := m.Validate()
	require.NoError(err)

	resp, err := th.SystemAdminClientPP.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   m,
		Replace:    true,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	assert.NoError(err)
	api4.CheckOKStatus(th.t, resp)

	resp, err = th.SystemAdminClientPP.InstallApp(m.AppID, apps.DeployHTTP)
	assert.NoError(err)
	api4.CheckOKStatus(th.t, resp)

	creq := apps.CallRequest{
		Context: apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID:     m.AppID,
				ChannelID: th.ServerTestHelper.BasicChannel.Id,
				TeamID:    th.ServerTestHelper.BasicTeam.Id,
			},
		},
		Call: apps.Call{
			Path: "/setup/user",
			Expand: &apps.Expand{
				ActingUserAccessToken: apps.ExpandAll,
			},
		},
	}

	cres, _, err := th.UserClientPP.Call(creq)
	assert.NoError(err)
	assert.NotNil(cres)
	assert.Equal(apps.CallResponseTypeOK, cres.Type)
	assert.Empty(cres.ErrorText)

	creq.Call.Path = "/setup/user2"

	cres, _, err = th.User2ClientPP.Call(creq)
	assert.NoError(err)
	assert.NotNil(cres)
	assert.Equal(apps.CallResponseTypeOK, cres.Type)
	assert.Empty(cres.ErrorText)

	require.NotNil(asBot)
	require.NotNil(asUser)
	require.NotNil(asUser2)

	return TestApp{
		Manifest: m,
		AsUser:   asUser.ClientPP,
		AsUser2:  asUser2.ClientPP,
		AsBot:    asBot.ClientPP,
	}
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
		f(t, th.UserClientPP)
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
