package restapitest

import (
	"bytes"
	"embed"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

func (th *Helper) InstallAppWithCleanup(app *goapp.App) *apps.App {
	installed := th.InstallApp(app)
	th.Cleanup(func() { th.UninstallApp(installed.AppID) })
	return installed
}

func (th *Helper) InstallApp(app *goapp.App) *apps.App {
	require := require.New(th)
	assert := assert.New(th)

	appServer := app.NewTestServer(th)
	th.Logf("started: '%s', listening on %s", app.Manifest.AppID, appServer.Listener.Addr().String())
	th.Cleanup(func() {
		appServer.Close()
		th.Logf("shut down: '%s'", app.Manifest.AppID)
	})
	require.Equal(appServer.URL, app.Manifest.Deploy.HTTP.RootURL)

	_, resp, err := th.SystemAdminClientPP.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   app.Manifest,
		Replace:    true,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	assert.NoError(err)
	api4.CheckOKStatus(th, resp)

	installed, resp, err := th.SystemAdminClientPP.InstallApp(app.Manifest.AppID, apps.DeployHTTP)
	assert.NoError(err)
	api4.CheckOKStatus(th, resp)

	th.Logf("installed: '%s' on '%s'", app.Manifest.AppID, app.Manifest.Deploy.HTTP.RootURL)
	return installed
}

func (th *Helper) DisableApp(app *goapp.App) {
	_, resp, err := th.SystemAdminClientPP.DisableApp(app.Manifest.AppID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("disabled: '%s'", app.Manifest.AppID)
}

func (th *Helper) EnableApp(app *goapp.App) {
	_, resp, err := th.SystemAdminClientPP.EnableApp(app.Manifest.AppID)
	require.NoError(th, err)
	api4.CheckOKStatus(th, resp)
	th.Logf("enabled: '%s'", app.Manifest.AppID)
}

func (th *Helper) UninstallApp(appID apps.AppID) {
	require := require.New(th)
	resp, err := th.SystemAdminClientPP.UninstallApp(appID)
	require.NoError(err)
	api4.CheckOKStatus(th, resp)
	th.Logf("uninstall: '%s'", appID)
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

	th.Logf("installed the apps plugin from bundle: '%s'", bundle)
}
