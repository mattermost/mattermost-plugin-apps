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

func (th *Helper) InstallAppWithCleanup(app *goapp.App) {
	th.InstallApp(app)
	th.Cleanup(func() { th.UninstallApp(app.Manifest.AppID) })
}

func (th *Helper) InstallApp(app *goapp.App) {
	require := require.New(th)
	assert := assert.New(th)

	appServer := app.NewTestServer()
	th.Cleanup(appServer.Close)
	require.Equal(appServer.URL, app.Manifest.Deploy.HTTP.RootURL)

	resp, err := th.SystemAdminClientPP.UpdateAppListing(appclient.UpdateAppListingRequest{
		Manifest:   app.Manifest,
		Replace:    true,
		AddDeploys: apps.DeployTypes{apps.DeployHTTP},
	})
	assert.NoError(err)
	api4.CheckOKStatus(th, resp)

	resp, err = th.SystemAdminClientPP.InstallApp(app.Manifest.AppID, apps.DeployHTTP)
	assert.NoError(err)
	api4.CheckOKStatus(th, resp)
}

func (th *Helper) UninstallApp(appID apps.AppID) {
	require := require.New(th)
	resp, err := th.SystemAdminClientPP.UninstallApp(appID)
	require.NoError(err)
	api4.CheckOKStatus(th, resp)
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
