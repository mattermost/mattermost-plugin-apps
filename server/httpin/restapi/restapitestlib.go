// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

//go:build rest_api_test

package restapi

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	LocalClientPP       *appclient.ClientPP
}

type TestClientPP struct {
	*appclient.ClientPP
	UserID string
}

type TestApp struct {
	t        *testing.T
	Manifest apps.Manifest
	AsUser   *TestClientPP
	AsUser2  *TestClientPP
	AsBot    *TestClientPP
}

func (th *TestHelper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func Setup(t *testing.T) *TestHelper {
	// Unset SiteURL, just in case it's set
	err := os.Unsetenv("MM_SERVICESETTINGS_SITEURL")
	require.NoError(t, err)

	// Enable Apps Feature Flags via env variables as it can't be set via the config
	t.Setenv("MM_FEATUREFLAGS_APPSENABLED", "true")

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

	userClient4 := th.ServerTestHelper.CreateClient()
	th.ServerTestHelper.LoginBasicWithClient(userClient4)
	th.UserClientPP = th.CreateClientPP()
	th.UserClientPP.AuthToken = userClient4.AuthToken
	th.UserClientPP.AuthType = userClient4.AuthType

	user2Client4 := th.ServerTestHelper.CreateClient()
	th.ServerTestHelper.LoginBasic2WithClient(user2Client4)
	th.User2ClientPP = th.CreateClientPP()
	th.User2ClientPP.AuthToken = user2Client4.AuthToken
	th.User2ClientPP.AuthType = user2Client4.AuthType

	systemAminClient4 := th.ServerTestHelper.CreateClient()
	th.ServerTestHelper.LoginSystemAdminWithClient(systemAminClient4)
	th.SystemAdminClientPP = th.CreateClientPP()
	th.SystemAdminClientPP.AuthToken = systemAminClient4.AuthToken
	th.SystemAdminClientPP.AuthType = systemAminClient4.AuthType

	th.LocalClientPP = th.CreateLocalClient(*th.ServerTestHelper.App.Config().ServiceSettings.LocalModeSocketLocation)

	return th
}

// Sets up the PP for test
func (th *TestHelper) SetupPP() {
	require := require.New(th.t)

	bundle := os.Getenv("PLUGIN_BUNDLE")
	require.NotEmpty(bundle, "PLUGIN_BUNDLE is not set, please run `make test-rest-api`")

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

func (th *TestHelper) CreateClientPP() *appclient.ClientPP {
	cfg := th.ServerTestHelper.App.Config()

	siteURL, err := url.Parse(*cfg.ServiceSettings.SiteURL)
	require.NoError(th.t, err)

	url := fmt.Sprintf("http://localhost:%v", th.ServerTestHelper.App.Srv().ListenAddr.Port) + siteURL.Path

	return appclient.NewAppsPluginAPIClient(url)
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

// Sets up the PP for test
func (th *TestHelper) SetupApp(m apps.Manifest) TestApp {
	th.t.Helper()
	require := require.New(th.t)
	assert := assert.New(th.t)

	var (
		asUser *appclient.Client
		userID string

		asUser2 *appclient.Client
		user2ID string

		asBot     *appclient.Client
		botUserID string
	)

	router := mux.NewRouter()
	router.HandleFunc(apps.DefaultPing.Path, httputils.DoHandleJSON(apps.NewDataResponse(nil)))
	router.HandleFunc("/setup/user", func(w http.ResponseWriter, r *http.Request) {
		creq, err := apps.CallRequestFromJSONReader(r.Body)
		require.NoError(err)
		require.NotNil(creq)

		asUser = appclient.AsActingUser(creq.Context)
		userID = creq.Context.ActingUser.Id

		asBot = appclient.AsBot(creq.Context)
		botUserID = creq.Context.BotUserID

		httputils.WriteJSON(w, apps.NewDataResponse(nil))
	})
	router.HandleFunc("/setup/user2", func(w http.ResponseWriter, r *http.Request) {
		creq, err := apps.CallRequestFromJSONReader(r.Body)
		require.NoError(err)
		require.NotNil(creq)

		asUser2 = appclient.AsActingUser(creq.Context)
		user2ID = creq.Context.ActingUser.Id

		httputils.WriteJSON(w, apps.NewDataResponse(nil))
	})
	appServer := httptest.NewServer(router)
	th.t.Cleanup(appServer.Close)

	m.HTTP = &apps.HTTP{
		RootURL: appServer.URL,
		UseJWT:  false,
	}
	m.HomepageURL = appServer.URL

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
				User:                  apps.ExpandID,
			},
		},
	}

	cres, resp, err := th.UserClientPP.Call(creq)
	assert.NoError(err)
	api4.CheckOKStatus(th.t, resp)
	assert.NotNil(cres)
	assert.Equal(apps.CallResponseTypeOK, cres.Type)
	assert.Empty(cres.Text)

	creq.Call.Path = "/setup/user2"

	cres, resp, err = th.User2ClientPP.Call(creq)
	assert.NoError(err)
	api4.CheckOKStatus(th.t, resp)
	assert.NotNil(cres)
	assert.Equal(apps.CallResponseTypeOK, cres.Type)
	assert.Empty(cres.Text)

	require.NotNil(asBot)
	require.NotNil(asUser)
	require.NotNil(asUser2)

	_, resp, err = th.ServerTestHelper.SystemAdminClient.AddTeamMember(th.ServerTestHelper.BasicTeam.Id, botUserID)
	assert.NoError(err)
	api4.CheckCreatedStatus(th.t, resp)

	_, resp, err = th.ServerTestHelper.SystemAdminClient.AddChannelMember(th.ServerTestHelper.BasicChannel.Id, botUserID)
	assert.NoError(err)
	api4.CheckCreatedStatus(th.t, resp)

	_, resp, err = th.ServerTestHelper.SystemAdminClient.AddChannelMember(th.ServerTestHelper.BasicChannel2.Id, botUserID)
	assert.NoError(err)
	api4.CheckCreatedStatus(th.t, resp)

	return TestApp{
		t:        th.t,
		Manifest: m,
		AsUser:   &TestClientPP{asUser.ClientPP, userID},
		AsUser2:  &TestClientPP{asUser2.ClientPP, user2ID},
		AsBot:    &TestClientPP{asBot.ClientPP, botUserID},
	}
}

func (ta TestApp) TestForUser(f func(*testing.T, *TestClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	ta.t.Run(testName+"AsUser", func(t *testing.T) {
		f(t, ta.AsUser)
	})
}

func (ta TestApp) TestForUser2(f func(*testing.T, *TestClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	ta.t.Run(testName+"AsUser2", func(t *testing.T) {
		f(t, ta.AsUser)
	})
}

func (ta TestApp) TestForBot(f func(*testing.T, *TestClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	ta.t.Run(testName+"AsBot", func(t *testing.T) {
		f(t, ta.AsBot)
	})
}

func (ta TestApp) TestForUserAndBot(f func(*testing.T, *TestClientPP), name ...string) {
	ta.TestForUser(f)
	ta.TestForBot(f)
}

func (ta TestApp) TestForTwoUsersAndBot(f func(*testing.T, *TestClientPP), name ...string) {
	ta.TestForUser(f)
	ta.TestForUser2(f)
	ta.TestForBot(f)
}
