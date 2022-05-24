// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	_ "embed"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type notifyApp struct {
	*goapp.App
	lastNotification apps.CallRequest
}

func newNotifyApp(th *Helper, appIDSuffix apps.AppID) *notifyApp {
	app := &notifyApp{
		App: goapp.MakeAppOrPanic(
			apps.Manifest{
				AppID:       apps.AppID("notify-") + appIDSuffix,
				Version:     "v1.0.0",
				DisplayName: "Returns bindings",
				HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
				RequestedPermissions: []apps.Permission{
					apps.PermissionActAsBot,
					apps.PermissionActAsUser,
				},
			},
		),
	}

	params := func(creq goapp.CallRequest) (*appclient.Client, apps.Subscription) {
		var sub apps.Subscription
		utils.Remarshal(&sub, creq.Values["sub"])
		require.NotEmpty(th, creq.Context.BotAccessToken)
		return appclient.AsBot(creq.Context), sub
	}

	app.HandleCall("/subscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			return respond("subscribed", client.Subscribe(&sub))
		})

	app.HandleCall("/unsubscribe",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, sub := params(creq)
			return respond("unsubscribed", client.Unsubscribe(&sub))
		})

	app.HandleCall("/notify",
		func(creq goapp.CallRequest) apps.CallResponse {
			app.lastNotification = creq.CallRequest
			return respond("OK", nil)
		})
}

func testNotify(th *Helper) {
	th.Run("User created deleted no expand", func(th *Helper) {
		app := newNotifyApp(th, "create-delete")
		th.InstallAppWithCleanup(app.App)
		require := require.New(th)

		appID := app.Manifest.AppID
		subject := apps.SubjectUserCreated
		expand := apps.Expand{}

		cresp := th.HappyCall(appID, apps.CallRequest{
			Call: *apps.NewCall("/subscribe"),
			Values: map[string]interface{}{
				"sub": apps.Subscription{
					Subject: subject,
					Call: *apps.NewCall("/notify").WithExpand(expand),
				},
			},
		})

		

		require.Equal(`subscribed`, cresp.Text)

	})

}
