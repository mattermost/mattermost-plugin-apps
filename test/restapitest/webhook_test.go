// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

const webhookAppID = apps.AppID("webhooktest")
const appURL = "/plugins/com.mattermost.apps/apps/" + string(webhookAppID)

func newWebhookApp(t *testing.T, onRemoteWebhook *apps.Call) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       webhookAppID,
			Version:     "v1.1.0",
			DisplayName: "tests App's Webhook APIs",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionRemoteWebhooks,
			},
			OnRemoteWebhook: onRemoteWebhook,
		},
	)

	return app
}

func sendWebhookHeadRequest(th *Helper, secret string) error {
	url := appURL + "/webhook?secret=" + secret

	r, err := th.UserClientPP.DoAPIHEAD(url, "")
	if err != nil && len(err.Error()) > 0 {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.Errorf("received error status code: %v", r.StatusCode)
	}

	return nil
}

func sendWebhookPostRequest(th *Helper, secret string, path string) error {
	url := appURL + path + "?secret=" + secret
	fmt.Println(url)

	r, err := th.UserClientPP.DoAPIPOST(url, "")
	if err != nil && len(err.Error()) > 0 {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.Errorf("received error status code: %v", r.StatusCode)
	}

	return nil
}

func getTestWebhookSecret(th *Helper, app *goapp.App) string {
	path := "/get_webhook_secret"

	// call handler to receive and return webhook secret
	app.HandleCall(path,
		func(creq goapp.CallRequest) apps.CallResponse {
			return respond(creq.Context.App.WebhookSecret, nil)
		})

	creq := apps.CallRequest{
		Call: *apps.NewCall(path).
			WithExpand(apps.Expand{App: apps.ExpandAll}),
	}

	cresp := th.HappyAdminCall(webhookAppID, creq)
	return cresp.Text
}

func testWebhookAuth(th *Helper) {
	app := newWebhookApp(th.T, nil)
	th.InstallAppWithCleanup(app)

	webhookSecret := getTestWebhookSecret(th, app)

	th.Run("webhook HEAD request with secret passes", func(th *Helper) {
		err := sendWebhookHeadRequest(th, webhookSecret)
		require.NoError(th, err)
	})

	th.Run("webhook HEAD request with invalid secret fails", func(th *Helper) {
		err := sendWebhookHeadRequest(th, webhookSecret+"nope")
		require.Error(th, err)
	})

	th.Run("webhook HEAD request with missing secret fails", func(th *Helper) {
		err := sendWebhookHeadRequest(th, "")
		require.Error(th, err)
	})

	th.Run("webhook POST request with secret passes", func(th *Helper) {
		err := sendWebhookPostRequest(th, webhookSecret, "/webhook")
		require.NoError(th, err)
	})

	th.Run("webhook POST request with invalid secret fails", func(th *Helper) {
		err := sendWebhookPostRequest(th, webhookSecret+"nope", "/webhook")
		require.Error(th, err)
	})

	th.Run("webhook POST request with missing secret fails", func(th *Helper) {
		err := sendWebhookPostRequest(th, "", "/webhook")
		require.Error(th, err)
	})
}

func testWebhookPaths(th *Helper) {
	for name, tc := range map[string]struct {
		onRemoteWebhook *apps.Call
		reqPath         string
		listenPath      string
		called          bool
	}{
		"without OnRemoteWebhook, default webhook path": {
			reqPath:    "/webhook",
			listenPath: "/webhook",
			called:     true,
		},
		"without OnRemoteWebhook, webhook path with suffix": {
			reqPath:    "/webhook/request-suffix",
			listenPath: "/webhook/request-suffix",
			called:     true,
		},
		"with OnRemoteWebhook, short request path": {
			onRemoteWebhook: apps.NewCall("/my-webhook"),
			reqPath:         "/webhook",
			listenPath:      "/my-webhook",
			called:          true,
		},
		"with OnRemoteWebhook, long request path": {
			onRemoteWebhook: apps.NewCall("/my-webhook"),
			reqPath:         "/webhook/request-suffix",
			listenPath:      "/my-webhook/request-suffix",
			called:          true,
		},
		"with OnRemoteWebhook, long OnRemoteWebhook path": {
			onRemoteWebhook: apps.NewCall("/my-webhook/manifest-suffix"),
			reqPath:         "/webhook",
			listenPath:      "/my-webhook/manifest-suffix",
			called:          true,
		},
		"with OnRemoteWebhook, long OnRemoteWebhook path request path": {
			onRemoteWebhook: apps.NewCall("/my-webhook/manifest-suffix"),
			reqPath:         "/webhook/request-suffix",
			listenPath:      "/my-webhook/manifest-suffix/request-suffix",
			called:          true,
		},
	} {
		th.Run(name, func(th *Helper) {
			app := newWebhookApp(th.T, tc.onRemoteWebhook)
			th.InstallAppWithCleanup(app)

			calledChan := make(chan bool)

			app.HandleCall(tc.listenPath, func(cr goapp.CallRequest) apps.CallResponse {
				calledChan <- true
				return apps.NewTextResponse("success")
			})

			webhookSecret := getTestWebhookSecret(th, app)
			err := sendWebhookPostRequest(th, webhookSecret, tc.reqPath)
			require.NoError(th, err)

			var result bool
			select {
			case result = <-calledChan:
			case <-time.After(time.Second * 1):
			}

			require.True(th, result)
		})
	}
}
