// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/v8/channels/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/server/builtin"
)

func testCalls(th *Helper) {
	addFailHandler := func(r *mux.Router, path string) {
		r.Path(path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			th.Errorf("%s should not be accessible to users", path)
		})
	}

	assertCallForbidden := func(path string) {
		creq := apps.CallRequest{
			Context: apps.Context{
				UserAgentContext: apps.UserAgentContext{
					AppID: "calls_test",
				},
			},
			Call: apps.Call{
				Path: path,
			},
		}

		cresp, resp, err := th.UserClientPP.Call(creq)
		require.NoError(th, err)
		api4.CheckOKStatus(th, resp)
		require.NotNil(th, cresp)
		assert.Equal(th, apps.CallResponseTypeError, cresp.Type)
		assert.Contains(th, cresp.Text, "forbidden call path")
	}

	th.Run("test forbidden paths with defaults", func(th *Helper) {
		app := goapp.MakeAppOrPanic(
			apps.Manifest{
				AppID:       "calls_test",
				DisplayName: "Test forbidden call paths",
				HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			},
		)

		th.InstallAppWithCleanup(app)

		failRouter := mux.NewRouter()
		app.Router = failRouter

		addFailHandler(app.Router, apps.DefaultPing.Path)
		addFailHandler(app.Router, apps.DefaultGetOAuth2ConnectURL.Path)
		addFailHandler(app.Router, apps.DefaultOnOAuth2Complete.Path)
		addFailHandler(app.Router, apps.DefaultOnRemoteWebhook.Path)

		assertCallForbidden(apps.DefaultPing.Path)
		assertCallForbidden(apps.DefaultGetOAuth2ConnectURL.Path)
		assertCallForbidden(apps.DefaultOnOAuth2Complete.Path)
		assertCallForbidden(apps.DefaultOnRemoteWebhook.Path)
	})

	th.Run("test forbidden paths with custom calls", func(th *Helper) {
		app := goapp.MakeAppOrPanic(
			apps.Manifest{
				AppID:               "calls_test",
				DisplayName:         "Test forbidden call paths",
				HomepageURL:         "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
				OnInstall:           &apps.Call{Path: "/prefix/on_install"},
				OnUninstall:         &apps.Call{Path: "/prefix/on_uninstall"},
				OnVersionChanged:    &apps.Call{Path: "/prefix/on_version_changed"},
				OnEnable:            &apps.Call{Path: "/prefix/on_enable"},
				OnDisable:           &apps.Call{Path: "/prefix/on_disable"},
				GetOAuth2ConnectURL: &apps.Call{Path: "/prefix/get_oauth2_connect_url"},
				OnOAuth2Complete:    &apps.Call{Path: "/prefix/on_oauth2_complete"},
				OnRemoteWebhook:     &apps.Call{Path: "/prefix/on_remote_webhook"},
			},
		)
		app.HandleCall("/prefix/on_install", func(cr goapp.CallRequest) apps.CallResponse { return apps.CallResponse{Type: apps.CallResponseTypeOK} })
		app.HandleCall("/prefix/on_uninstall", func(cr goapp.CallRequest) apps.CallResponse { return apps.CallResponse{Type: apps.CallResponseTypeOK} })
		th.InstallAppWithCleanup(app)

		originalRouter := app.Router

		failRouter := mux.NewRouter()

		addFailHandler(failRouter, "/prefix/on_install")
		addFailHandler(failRouter, "/prefix/on_uninstall")
		addFailHandler(failRouter, "/prefix/on_version_changed")
		addFailHandler(failRouter, "/prefix/on_enable")
		addFailHandler(failRouter, "/prefix/on_disable")
		addFailHandler(failRouter, "/prefix/get_oauth2_connect_url")
		addFailHandler(failRouter, "/prefix/on_oauth2_complete")
		addFailHandler(failRouter, "/prefix/on_remote_webhook")

		app.Router = failRouter

		assertCallForbidden("/prefix/on_install")
		assertCallForbidden("/prefix/on_uninstall")
		assertCallForbidden("/prefix/on_version_changed")
		assertCallForbidden("/prefix/on_enable")
		assertCallForbidden("/prefix/on_disable")
		assertCallForbidden("/prefix/get_oauth2_connect_url")
		assertCallForbidden("/prefix/on_oauth2_complete")
		assertCallForbidden("/prefix/on_remote_webhook")

		// Revert back to the original router for cleanup
		app.Router = originalRouter
	})

	th.Run("user can not invoke builtin debug calls", func(th *Helper) {
		infoRequest := apps.CallRequest{
			Call: *apps.NewCall(builtin.PathDebugKVInfo).WithExpand(apps.Expand{
				ActingUser: apps.ExpandSummary,
			}),
			Values: map[string]interface{}{
				builtin.FieldAppID: uninstallID,
			},
		}

		cresp, _, err := th.Call(builtin.AppID, infoRequest)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Regexp(th, `user \w+ \(\w+\) is not a sysadmin: unauthorized`, cresp.Text)
	})
}
