// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
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

var testOAuth2User = map[string]interface{}{
	"test_bool":   true,
	"test_string": "test",
}

func newOAuth2App(t *testing.T) *goapp.App {
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

	largeJSON := struct {
		Fields []interface{}
	}{}
	for {
		data, err := json.Marshal(largeJSON)
		require.NoError(t, err)
		if len(data) > httpin.MaxKVStoreValueLength {
			break
		}
		largeJSON.Fields = append(largeJSON.Fields, largeJSON)
	}

	app.HandleCall("/get-user",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			v := map[string]interface{}{}
			err := client.GetOAuth2User(&v)
			return respond(utils.ToJSON(v), err)
		})

	app.HandleCall("/store-user",
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

	app.HandleCall("/err-user-too-large",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			err := client.StoreOAuth2User(largeJSON)
			return respond("impossible", err)
		})

	app.HandleCall("/err-user-not-json",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			resp, err := client.DoAPIPOST(
				client.GetPluginRoute(appclient.AppsPluginName)+appspath.API+appspath.OAuth2User,
				"test") // nolint:bodyclose
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			return respond("impossible", err)
		})

	app.HandleCall("/err-app-too-large",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			err := client.StoreOAuth2App(apps.OAuth2App{
				ClientID:      "test_id",
				ClientSecret:  "test_secret",
				RemoteRootURL: "test.test",
				Data:          largeJSON,
			})
			return respond("impossible", err)
		})

	app.HandleCall("/err-app-not-json",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, _ := params(creq)
			_, err := client.DoAPIPOST( // nolint:bodyclose
				client.GetPluginRoute(appclient.AppsPluginName)+appspath.API+appspath.OAuth2App,
				"test")
			return respond("impossible", err)
		})

	app.HandleCall("/echo", Echo)

	return app
}

func oauth2Call(th *Helper, path string, value interface{}) *apps.CallResponse {
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
	return th.HappyCall(oauth2ID, creq)
}

func testOAuth2(th *Helper) {
	th.InstallAppWithCleanup(newOAuth2App(th.T))

	storeOAuth2App := func(th *Helper, oapp apps.OAuth2App) {
		th.Helper()
		creq := apps.CallRequest{
			Call: *apps.NewCall("/store-app").WithExpand(apps.Expand{
				ActingUser:            apps.ExpandSummary.Required(),
				ActingUserAccessToken: apps.ExpandAll,
			}),
			Values: model.StringInterface{
				"as_bot": false,
				"value":  oapp,
			},
		}
		// By making an admin call, the admin-level token should propagage to the app
		cresp := th.HappyAdminCall(oauth2ID, creq)
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
		require.Equal(th, `stored`, cresp.Text)
	}

	cleanupOAuth2App := func(th *Helper) func() {
		return func() {
			storeOAuth2App(th, apps.OAuth2App{})
		}
	}

	cleanupOAuth2User := func(th *Helper) func() {
		return func() {
			_ = oauth2Call(th, "/store-user", struct{}{})
			cresp := oauth2Call(th, "/get-user", nil)
			require.Equal(th, `{}`, cresp.Text)
		}
	}

	th.Run("users can store and get OAuth2User via REST API", func(th *Helper) {
		th.Cleanup(cleanupOAuth2User(th))

		cresp := oauth2Call(th, "/store-user", testOAuth2User)
		require.Equal(th, `stored`, cresp.Text)

		cresp = oauth2Call(th, "/get-user", nil)
		require.Equal(th, `{"test_bool":true,"test_string":"test"}`, cresp.Text)
	})

	th.Run("System administrators can store OAuth2App", func(th *Helper) {
		th.Cleanup(cleanupOAuth2App(th))
		storeOAuth2App(th, testOAuth2App)
	})

	th.Run("User and bot calls can expand OAuth2App", func(th *Helper) {
		// th.Skip("https://mattermost.atlassian.net/browse/MM-48448")
		th.Cleanup(func() {
			cleanupOAuth2App(th)()
			cleanupOAuth2User(th)()
		})

		// Store the app and the user.
		storeOAuth2App(th, testOAuth2App)
		cresp := oauth2Call(th, "/store-user", testOAuth2User)
		require.Equal(th, `stored`, cresp.Text)

		// Call echo and verify the expand result.
		cresp = oauth2Call(th, "/echo", nil)
		creq := apps.CallRequest{}
		require.Equal(th, apps.CallResponseTypeOK, cresp.Type)
		err := json.Unmarshal([]byte(cresp.Text), &creq)
		require.NoError(th, err)
		require.EqualValues(th, &testOAuth2App, &creq.Context.ExpandedContext.OAuth2.OAuth2App)
		require.EqualValues(th, map[string]interface{}{"test_bool": true, "test_string": "test"}, creq.Context.ExpandedContext.OAuth2.User)
	})

	th.Run("Error unauthenticated requests are rejected", func(th *Helper) {
		client := th.CreateUnauthenticatedClientPP()

		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}
		resp, err := client.StoreOAuth2User(in)
		require.Error(th, err)
		api4.CheckUnauthorizedStatus(th, resp)

		var out map[string]interface{}
		resp, err = client.GetOAuth2User(&out)
		require.Error(th, err)
		api4.CheckUnauthorizedStatus(th, resp)

		resp, err = client.StoreOAuth2App(testOAuth2App)
		require.Error(th, err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("Error StoreOAuth2User is size limited", func(th *Helper) {
		th.Cleanup(cleanupOAuth2User(th))

		// set a "previous" value.
		cresp := oauth2Call(th, "/store-user", map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		})
		require.Equal(th, `stored`, cresp.Text)

		creq := apps.CallRequest{
			Call: *apps.NewCall("/err-user-too-large").
				WithExpand(apps.Expand{
					ActingUser:            apps.ExpandSummary.Required(),
					ActingUserAccessToken: apps.ExpandAll.Required(),
				}),
		}
		cresp, resp, err := th.Call(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Equal(th, `size limit of 8Kb exceeded`, cresp.Text)
		api4.CheckOKStatus(th, resp)

		// verify "previous" value unchanged.
		cresp = oauth2Call(th, "/get-user", nil)
		require.Equal(th, `{"test_bool":true,"test_string":"test"}`, cresp.Text)
	})

	th.Run("Error StoreOAuth2User requires JSON", func(th *Helper) {
		th.Cleanup(cleanupOAuth2User(th))

		// set a "previous" value.
		cresp := oauth2Call(th, "/store-user", map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		})

		require.Equal(th, `stored`, cresp.Text)
		creq := apps.CallRequest{
			Call: *apps.NewCall("/err-user-not-json").
				WithExpand(apps.Expand{
					ActingUser:            apps.ExpandSummary.Required(),
					ActingUserAccessToken: apps.ExpandAll.Required(),
				}),
		}
		cresp, resp, err := th.Call(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, `payload is not valid JSON: invalid input`, cresp.Text)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		api4.CheckOKStatus(th, resp)

		// verify "previous" value unchanged.
		cresp = oauth2Call(th, "/get-user", nil)
		require.Equal(th, `{"test_bool":true,"test_string":"test"}`, cresp.Text)
	})

	th.Run("Error StoreOAuth2App is size limited", func(th *Helper) {
		th.Cleanup(cleanupOAuth2App(th))

		creq := apps.CallRequest{
			Call: *apps.NewCall("/err-app-too-large").
				WithExpand(apps.Expand{
					ActingUser:            apps.ExpandSummary.Required(),
					ActingUserAccessToken: apps.ExpandAll.Required(),
				}),
		}
		cresp, resp, err := th.AdminCall(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Equal(th, `size limit of 8Kb exceeded`, cresp.Text)
		api4.CheckOKStatus(th, resp)
	})

	th.Run("Error StoreOAuth2App requires JSON", func(th *Helper) {
		th.Cleanup(cleanupOAuth2App(th))

		creq := apps.CallRequest{
			Call: *apps.NewCall("/err-app-not-json").
				WithExpand(apps.Expand{
					ActingUser:            apps.ExpandSummary.Required(),
					ActingUserAccessToken: apps.ExpandAll.Required(),
				}),
		}
		cresp, resp, err := th.AdminCall(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, `OAuth2App is not valid JSON: invalid character 'e' in literal true (expecting 'r'): invalid input`, cresp.Text)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		api4.CheckOKStatus(th, resp)
	})

	th.Run("Error Bots have no access to OAuth2User via REST API", func(th *Helper) {
		// try to get.
		creq := apps.CallRequest{
			Call: *apps.NewCall("/get-user"),
			Values: model.StringInterface{
				"as_bot": true,
			},
		}
		cresp, resp, err := th.Call(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Equal(th, `@oauth2test (tests App's OAuth2 APIs): is a bot: unauthorized`, cresp.Text)
		// TODO: should be a 403!
		api4.CheckOKStatus(th, resp)

		// try to store.
		creq.Call = *apps.NewCall("/store-user")
		creq.Values["value"] = map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}
		cresp, resp, err = th.Call(oauth2ID, creq)
		require.NoError(th, err)
		require.Equal(th, apps.CallResponseTypeError, cresp.Type)
		require.Equal(th, `@oauth2test (tests App's OAuth2 APIs): is a bot: unauthorized`, cresp.Text)
		require.NotNil(th, resp)
		// TODO: should be a 403!
		require.Equal(th, resp.StatusCode, 200)
	})

	th.Run("Error Users and bots can not store OAuth2App", func(th *Helper) {
		for _, asBot := range []bool{true, false} {
			name := "as acting user"
			if asBot {
				name = "as bot"
			}
			th.Run(name, func(th *Helper) {
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
				require.NoError(th, err)
				require.Equal(th, apps.CallResponseTypeError, cresp.Type)
				require.Equal(th, `access to this operation is limited to users with permission: manage_system: unauthorized`, cresp.Text)
				require.NotNil(th, resp)
				// TODO: should be a 403!
				require.Equal(th, resp.StatusCode, 200)
			})
		}
	})
}
