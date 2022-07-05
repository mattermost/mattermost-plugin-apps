package restapitest

import (
	"bytes"
	"embed"
	"os"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

func (th *Helper) InstallAppWithCleanup(app *goapp.App) {
	installed := th.InstallApp(app)
	th.LastInstalledApp = installed
	th.Cleanup(func() {
		_, _ = th.SystemAdminClientPP.UninstallApp(installed.AppID)
		th.Logf("uninstalled: '%s'", installed.AppID)
	})

	var appErr *model.AppError
	th.LastInstalledBotUser, appErr = th.ServerTestHelper.App.GetUser(th.LastInstalledApp.BotUserID)
	require.Nil(th, appErr)

	th.asBot = appClient{
		name:               "bot",
		expectedActingUser: th.LastInstalledBotUser,
		happyCall:          th.HappyCall,
		call:               th.Call,
		appActsAsBot:       true,
	}

	th.asUser = appClient{
		name:               "user",
		expectedActingUser: th.ServerTestHelper.BasicUser,
		happyCall:          th.HappyCall,
		call:               th.Call,
	}

	th.asUser2 = appClient{
		name:               "user2",
		expectedActingUser: th.ServerTestHelper.BasicUser2,
		happyCall:          th.HappyUser2Call,
		call:               th.User2Call,
	}

	th.asAdmin = appClient{
		name:                 "admin",
		expectedActingUser:   th.ServerTestHelper.SystemAdminUser,
		happyCall:            th.HappyAdminCall,
		call:                 th.AdminCall,
		appActsAsSystemAdmin: true,
	}
}

func (th *Helper) InstallApp(app *goapp.App) *apps.App {
	appServer := app.NewTestServer(th)
	th.Logf("started: '%s', listening on %s", app.Manifest.AppID, appServer.Listener.Addr().String())
	th.Cleanup(func() {
		appServer.Close()
		th.Logf("shut down: '%s'", app.Manifest.AppID)
	})
	require.Equal(th, appServer.URL, app.Manifest.Deploy.HTTP.RootURL)

	_, resp, err := th.SystemAdminClientPP.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   app.Manifest,
		Replace:    true,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)

	installed, resp, err := th.SystemAdminClientPP.InstallApp(app.Manifest.AppID, apps.DeployHTTP)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)

	th.Logf("installed: '%s' on '%s'", app.Manifest.AppID, app.Manifest.Deploy.HTTP.RootURL)
	return installed
}

func (th *Helper) DisableApp(app *goapp.App) {
	resp, err := th.SystemAdminClientPP.DisableApp(app.Manifest.AppID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("disabled: '%s'", app.Manifest.AppID)
}

func (th *Helper) EnableApp(app *goapp.App) {
	resp, err := th.SystemAdminClientPP.EnableApp(app.Manifest.AppID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("enabled: '%s'", app.Manifest.AppID)
}

func (th *Helper) UninstallApp(appID apps.AppID) {
	resp, err := th.SystemAdminClientPP.UninstallApp(appID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("uninstalled: '%s'", appID)
}

func (th *Helper) InstallAppsPlugin() {
	bundle := os.Getenv("PLUGIN_BUNDLE")
	require.NotEmpty(th, bundle, "PLUGIN_BUNDLE is not set, please run `make test-rest-api`")

	// Install the PP and enable it
	pluginBytes, err := os.ReadFile(bundle)
	require.NoError(th, err)
	require.NotNil(th, pluginBytes)

	manifest, appErr := th.ServerTestHelper.App.InstallPlugin(bytes.NewReader(pluginBytes), true)
	require.Nil(th, appErr)
	require.Equal(th, pluginID, manifest.Id)

	appErr = th.ServerTestHelper.App.EnablePlugin(pluginID)
	require.Nil(th, appErr)

	th.Logf("installed the apps plugin from bundle: '%s'", bundle)
}
