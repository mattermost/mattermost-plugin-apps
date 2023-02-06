// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed" // a test package, effectively
	"fmt"
	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

//go:embed static/icon.png
var iconPNG []byte

func testStatic(th *Helper) {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       "static_test",
			DisplayName: "Test static call paths",
			Icon:        "icon.png",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
		},
		goapp.WithStatic(static),
	)

	th.InstallAppWithCleanup(app)

	iconURL := fmt.Sprintf("/plugins/com.mattermost.apps/apps/%s/static/icon.png", app.Manifest.AppID)

	th.Run("static icon accessible as user", func(th *Helper) {
		resp, err := th.UserClientPP.DoAPIGET(iconURL, "")
		require.NoError(th, err)
		require.NotNil(th, resp)
		th.Cleanup(func() { _ = resp.Body.Close() })

		data, err := io.ReadAll(resp.Body)
		require.NoError(th, err)
		require.Equal(th, iconPNG, data)
	})

	th.Run("static icon accessible as user if app is disabled", func(th *Helper) {
		mmResp, err := th.SystemAdminClientPP.DisableApp(app.Manifest.AppID)
		assert.NoError(th, err)
		api4.CheckOKStatus(th, mmResp)

		th.Cleanup(func() {
			mmResp, err = th.SystemAdminClientPP.EnableApp(app.Manifest.AppID)
			assert.NoError(th, err)
			api4.CheckOKStatus(th, mmResp)
		})

		resp, err := th.UserClientPP.DoAPIGET(iconURL, "")
		require.NoError(th, err)
		require.NotNil(th, resp)
		th.Cleanup(func() { _ = resp.Body.Close() })

		data, err := io.ReadAll(resp.Body)
		require.NoError(th, err)
		require.Equal(th, iconPNG, data)
	})
}
