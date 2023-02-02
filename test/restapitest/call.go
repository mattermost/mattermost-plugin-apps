package restapitest

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

func testCalls(th *Helper) {
	th.Run("test forbidden paths", func(th *Helper) {
		app := goapp.MakeAppOrPanic(
			apps.Manifest{
				AppID:               "calls_test",
				DisplayName:         "Test forbidden call paths",
				HomepageURL:         "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
				OnInstall:           &apps.Call{Path: "/on_install"},
				OnUninstall:         &apps.Call{Path: "/on_uninstall"},
				OnVersionChanged:    &apps.Call{Path: "/on_version_changed"},
				OnEnable:            &apps.Call{Path: "/on_enable"},
				OnDisable:           &apps.Call{Path: "/on_disable"},
				GetOAuth2ConnectURL: &apps.Call{Path: "/get_oauth2_connect_url"},
				OnOAuth2Complete:    &apps.Call{Path: "/on_oauth2_complete"},
				OnRemoteWebhook:     &apps.Call{Path: "/on_remote_webhook"},
			},
		)
		app.HandleCall("/on_install", func(cr goapp.CallRequest) apps.CallResponse { return apps.CallResponse{Type: apps.CallResponseTypeOK} })
		app.HandleCall("/on_uninstall", func(cr goapp.CallRequest) apps.CallResponse { return apps.CallResponse{Type: apps.CallResponseTypeOK} })
		th.InstallAppWithCleanup(app)

		originalRouter := app.Router

		failRouter := mux.NewRouter()
		addFailHandler := func(path string) {
			failRouter.Path(path).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				th.Errorf("%s should not be accessible to users", path)
			})
		}
		addFailHandler("/on_install")
		addFailHandler("/on_uninstall")
		addFailHandler("/on_version_changed")
		addFailHandler("/on_enable")
		addFailHandler("/on_disable")
		addFailHandler("/get_oauth2_connect_url")
		addFailHandler("/on_oauth2_complete")
		addFailHandler("/on_remote_webhook")

		app.Router = failRouter

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

		assertCallForbidden("/on_install")
		assertCallForbidden("/on_uninstall")
		assertCallForbidden("/on_version_changed")
		assertCallForbidden("/on_enable")
		assertCallForbidden("/on_disable")
		assertCallForbidden("/get_oauth2_connect_url")
		assertCallForbidden("/on_oauth2_complete")
		assertCallForbidden("/on_remote_webhook")

		// Revert back to the original router for cleanup
		app.Router = originalRouter
	})
}
