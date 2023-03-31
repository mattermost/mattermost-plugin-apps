package restapitest

import (
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

func testRenameBot(th *Helper) {
	initialDisplayName := "Initial DisplayName"
	manifest := apps.Manifest{
		AppID:       "renameBotTestApp",
		Version:     "v1.2.0",
		DisplayName: initialDisplayName,
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
	}

	// Install app with initial DisplayName
	app := goapp.MakeAppOrPanic(manifest)
	th.InstallAppWithCleanup(app)
	require.Equal(th, initialDisplayName, th.LastInstalledBotUser.GetDisplayName(model.ShowFullName))

	// Reinstall app with modified DisplayName
	th.UninstallApp(app.Manifest.AppID)
	modifiedDisplayName := "Modified DisplayName"
	app.Manifest.DisplayName = modifiedDisplayName
	th.InstallAppWithCleanup(app)
	require.Equal(th, modifiedDisplayName, th.LastInstalledBotUser.GetDisplayName(model.ShowFullName))
}
