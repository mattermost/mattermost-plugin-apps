package restapitest

import (
	"embed"

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
