// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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

	// Create the helper and register for cleanup.
	th := &Helper{
		T:                t,
		ServerTestHelper: serverTestHelper,
	}
	t.Cleanup(th.TearDown)

	th.InitClients()
	th.InstallAppsPlugin()
	for _, a := range apps {
		th.InstallAppWithCleanup(a)
	}
	return th
}

func (th *Helper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func (th *Helper) Run(name string, f func(th *Helper)) bool {
	return th.T.Run(name, func(t *testing.T) {
		h := *th
		h.T = t
		f(&h)
	})
}

func (th *Helper) Cleanup(f func()) {
	th.Helper()
	ss := strings.Split(th.Name(), "/")
	th.T.Cleanup(func() {
		th.T.Run("cleanup "+ss[len(ss)-1], func(*testing.T) { f() })
	})
}

func (th *Helper) NamedCleanup(name string, f func()) {
	th.Helper()
	th.T.Cleanup(func() {
		th.T.Run("cleanup "+name, func(*testing.T) {
			f()
		})
	})
}

func respond(text string, err error) apps.CallResponse {
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(text)
}

func (th *Helper) EqualBindings(expectedJSON []byte, actual []apps.Binding) {
	th.Helper()

	expected := []apps.Binding{}
	buf := &bytes.Buffer{}
	appsURL := fmt.Sprintf("http://localhost:%v/plugins/com.mattermost.apps/apps", th.ServerTestHelper.Server.ListenAddr.Port)
	template.Must(template.New("test").Parse(string(expectedJSON))).Execute(buf, struct {
		AppsURL string
	}{appsURL})
	err := json.NewDecoder(buf).Decode(&expected)
	require.NoError(th, err)
	require.EqualValues(th, utils.Pretty(expected), utils.Pretty(actual))
}
