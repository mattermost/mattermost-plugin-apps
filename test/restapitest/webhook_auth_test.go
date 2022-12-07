// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

const webhookAppID = apps.AppID("webhooktest")
const webhookURL = "/plugins/com.mattermost.apps/apps/" + string(webhookAppID) + "/webhook"

func newWebhookApp(t *testing.T) *goapp.App {
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
		},
	)

	return app
}

func sendWebhookHeadRequest(th *Helper, secret string) error {
	url := webhookURL + "?secret=" + secret

	r, err := th.UserClientPP.DoAPIHEAD(url, "")
	if err != nil && len(err.Error()) > 0 {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return errors.Errorf("received error status code: %v", r.StatusCode)
	}

	return nil
}

func sendWebhookPostRequest(th *Helper, secret string, body []byte) error {
	url := webhookURL + "?secret=" + secret

	r, err := th.UserClientPP.DoAPIPOST(url, string(body))
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

func testWebhook(th *Helper) {
	app := newWebhookApp(th.T)
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
		err := sendWebhookPostRequest(th, webhookSecret, nil)
		require.NoError(th, err)
	})

	th.Run("webhook POST request with invalid secret fails", func(th *Helper) {
		err := sendWebhookPostRequest(th, webhookSecret+"nope", nil)
		require.Error(th, err)
	})

	th.Run("webhook POST request with missing secret fails", func(th *Helper) {
		err := sendWebhookPostRequest(th, "", nil)
		require.Error(th, err)
	})
}
