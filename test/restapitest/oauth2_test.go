// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const oauth2ID = apps.AppID("oauth2test")

var testOAuth2App = apps.OAuth2App{
	ClientID:      "client-id",
	ClientSecret:  "client-secret",
	RemoteRootURL: "http://test.test/test",
	Data: map[string]interface{}{
		"test_bool":   true,
		"test_string": "test",
	},
}

func oauth2App(t *testing.T) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       oauth2ID,
			Version:     "v1.1.0",
			DisplayName: "tests App's OAuth2 APIs",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
				apps.PermissionRemoteOAuth2,
			},
		},
	)

	params := func(creq goapp.CallRequest) (client *appclient.Client, value interface{}) {
		asBot, _ := creq.BoolValue("as_bot")
		value = model.StringInterface{
			"test-name": "test-data",
		}
		if len(creq.Values) > 0 && creq.Values["value"] != nil {
			value = creq.Values["value"]
		}
		if asBot {
			require.NotEmpty(t, creq.Context.BotAccessToken)
			return appclient.AsBot(creq.Context), value
		}
		require.NotEmpty(t, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), value
	}

	app.HandleCall("/get",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			v := map[string]interface{}{}
			err := client.GetOAuth2User(&v)
			return respond(utils.ToJSON(v), err)
		})

	app.HandleCall("/store",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, value := params(creq)
			err := client.StoreOAuth2User(value)
			return respond("stored", err)
		})

	app.HandleCall("/store-app",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, value := params(creq)
			if value == nil {
				value = map[string]interface{}{}
			}
			oapp := apps.OAuth2App{}
			utils.Remarshal(&oapp, value)
			err := client.StoreOAuth2App(oapp)
			return respond("stored", err)
		})

	app.HandleCall("/echo", Echo)

	return app
}

func oauth2Call(th *Helper, path string, value interface{}) apps.CallResponse {
	creq := apps.CallRequest{
		Call: *apps.NewCall(path).
			WithExpand(apps.Expand{
				OAuth2App:  apps.ExpandAll,
				OAuth2User: apps.ExpandAll,
			}),
	}
	creq.Call.Expand.ActingUser = apps.ExpandSummary
	creq.Call.Expand.ActingUserAccessToken = apps.ExpandAll
	if value != nil {
		creq.Values = map[string]interface{}{
			"value": value,
		}
	}
	return *th.HappyCall(oauth2ID, creq)
}

func testOAuth2(th *Helper) {
	th.InstallApp(oauth2App(th.T))

	th.Run("Unauthenticated requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.CreateUnauthenticatedClientPP()

		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}
		resp, err := client.StoreOAuth2User(in)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		var out map[string]interface{}
		resp, err = client.GetOAuth2User(&out)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)

		resp, err = client.StoreOAuth2App(testOAuth2App)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("Users can store and get OAuth2User via REST API", func(th *Helper) {
		require := require.New(th)

		cresp := oauth2Call(th, "/get", nil)
		require.Equal(`{}`, cresp.Text)

		cresp = oauth2Call(th, "/store", map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		})
		require.Equal(`stored`, cresp.Text)

		cresp = oauth2Call(th, "/get", nil)
		require.Equal(`{"test_bool":true,"test_string":"test"}`, cresp.Text)
	})

	th.Run("Bots have no access to OAuth2User via REST API", func(th *Helper) {
		require := require.New(th)

		// try to get.
		creq := apps.CallRequest{
			Call: *apps.NewCall("/get"),
			Values: model.StringInterface{
				"as_bot": true,
			},
		}
		cresp, resp, err := th.Call(oauth2ID, creq)
		require.NoError(err)
		require.Equal(apps.CallResponseTypeError, cresp.Type)
		require.Equal(`tests App's OAuth2 APIs: is a bot`, cresp.Text)
		require.NotNil(resp)
		// TODO: should be a 403!
		require.Equal(resp.StatusCode, 200)

		// try to store.
		creq.Call = *apps.NewCall("/store")
		creq.Values["value"] = map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}
		cresp, resp, err = th.Call(oauth2ID, creq)
		require.NoError(err)
		require.Equal(apps.CallResponseTypeError, cresp.Type)
		require.Equal(`tests App's OAuth2 APIs: is a bot`, cresp.Text)
		require.NotNil(resp)
		// TODO: should be a 403!
		require.Equal(resp.StatusCode, 200)
	})

	th.Run("Users and bots can not store OAuth2App", func(th *Helper) {
		for _, asBot := range []bool{true, false} {
			name := "as acting user"
			if asBot {
				name = "as bot"
			}
			th.Run(name, func(th *Helper) {
				require := require.New(th)
				cresp, resp, err := th.Call(oauth2ID, apps.CallRequest{
					Call: *apps.NewCall("/store-app").WithExpand(apps.Expand{
						ActingUser:            apps.ExpandAll,
						ActingUserAccessToken: apps.ExpandAll,
					}),
					Values: model.StringInterface{
						"as_bot": asBot,
						"value":  testOAuth2App,
					},
				})
				require.NoError(err)
				require.Equal(apps.CallResponseTypeError, cresp.Type)
				require.Equal(`user is not a system admin: unauthorized`, cresp.Text)
				require.NotNil(resp)
				// TODO: should be a 403!
				require.Equal(resp.StatusCode, 200)
			})
		}
	})

	storeAppAsAdmin := func(th *Helper, oapp apps.OAuth2App) {
		th.Helper()
		require := require.New(th)
		creq := apps.CallRequest{
			Call: *apps.NewCall("/store-app").WithExpand(apps.Expand{
				ActingUser:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
			}),
			Values: model.StringInterface{
				"as_bot": false,
				"value":  oapp,
			},
		}
		// By making an admin call, the admin-level token should propagage to the app
		cresp := th.HappyAdminCall(oauth2ID, creq)
		require.Equal(apps.CallResponseTypeOK, cresp.Type)
		require.Equal(`stored`, cresp.Text)
	}

	th.Run("System administrators can store OAuth2App", func(th *Helper) {
		th.Cleanup(func() {
			storeAppAsAdmin(th, apps.OAuth2App{})
		})
		storeAppAsAdmin(th, testOAuth2App)
	})

	th.Run("User and bot calls can expand OAuth2App", func(th *Helper) {
		th.Skip("https://mattermost.atlassian.net/browse/MM-48448")
		th.Cleanup(func() {
			storeAppAsAdmin(th, apps.OAuth2App{})
		})
		storeAppAsAdmin(th, testOAuth2App)

		require := require.New(th)
		cresp := oauth2Call(th, "/echo", nil)

		creq := apps.CallRequest{}
		require.Equal(apps.CallResponseTypeOK, cresp.Type)
		err := json.Unmarshal([]byte(cresp.Text), &creq)
		require.NoError(err)
		require.EqualValues(&testOAuth2App, &creq.Context.ExpandedContext.OAuth2.OAuth2App)
		require.EqualValues(map[string]interface{}{"test_bool": true, "test_string": "test"}, creq.Context.ExpandedContext.OAuth2.User)
	})
}
