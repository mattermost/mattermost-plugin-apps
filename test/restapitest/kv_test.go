// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
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

const kvID = apps.AppID("kvtest")

func newKVApp(t testing.TB) *goapp.App {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       kvID,
			Version:     "v1.0.0",
			DisplayName: "tests access to the KV store",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
	)

	params := func(creq goapp.CallRequest) (client *appclient.Client, key string, value interface{}) {
		key, ok := creq.StringValue("key")
		if !ok {
			key = "test-key"
		}
		asBot, _ := creq.BoolValue("as_bot")
		value = model.StringInterface{
			"test-name": "test-data",
		}
		if len(creq.Values) > 0 && creq.Values["value"] != nil {
			value = creq.Values["value"]
		}
		if asBot {
			require.NotEmpty(t, creq.Context.BotAccessToken)
			return appclient.AsBot(creq.Context), key, value
		}
		require.NotEmpty(t, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), key, value
	}

	app.HandleCall("/get",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, key, _ := params(creq)
			v := map[string]interface{}{}
			err := client.KVGet("p", key, &v)
			require.NoError(t, err)
			return apps.NewTextResponse(utils.ToJSON(v))
		})

	app.HandleCall("/set",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, key, value := params(creq)
			changed, err := client.KVSet("p", key, value)
			require.NoError(t, err)
			return apps.NewTextResponse("%v", changed)
		})

	app.HandleCall("/delete",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, key, _ := params(creq)
			err := client.KVDelete("p", key)
			require.NoError(t, err)
			return apps.NewTextResponse("deleted")
		})

	return app
}

func kvCall(th *Helper, path string, asBot bool, key string, value interface{}) apps.CallResponse {
	creq := apps.CallRequest{
		Call: *apps.NewCall(path),
		Values: model.StringInterface{
			"as_bot": asBot,
		},
	}
	if !asBot {
		creq.Call.Expand = &apps.Expand{
			ActingUser:            apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
		}
	}
	if key != "" {
		creq.Values["key"] = key
	}
	if value != nil {
		creq.Values["value"] = value
	}
	return *th.HappyCall(kvID, creq)
}

func testKV(th *Helper) {
	th.InstallApp(newKVApp(th.T))

	th.Run("Unauthenticated requests are rejected", func(th *Helper) {
		assert := assert.New(th)
		client := th.CreateUnauthenticatedClientPP()

		changed, resp, err := client.KVSet("p", "id", "data")
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
		assert.False(changed)

		var out map[string]interface{}
		resp, err = client.KVGet("p", "id", &out)
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
		assert.False(changed)

		resp, err = client.KVDelete("p", "id")
		assert.Error(err)
		api4.CheckUnauthorizedStatus(th, resp)
	})

	th.Run("get-set-get-delete-get", func(th *Helper) {
		for _, asBot := range []bool{true, false} {
			name := "as acting user"
			if asBot {
				name = "as bot"
			}
			th.Run(name, func(th *Helper) {
				require := require.New(th)

				// Check that does not yet exist, should comes back empty.
				cresp := kvCall(th, "/get", asBot, "", nil)
				require.Equal(`{}`, cresp.Text)

				// Set.
				cresp = kvCall(th, "/set", asBot, "", nil)
				require.Equal("true", cresp.Text)

				// Check.
				cresp = kvCall(th, "/get", asBot, "", nil)
				require.Equal(`{"test-name":"test-data"}`, cresp.Text)

				// Delete.
				cresp = kvCall(th, "/delete", asBot, "", nil)
				require.Equal(`deleted`, cresp.Text)

				// Check again - deleted comes back as empty.
				cresp = kvCall(th, "/get", asBot, "", nil)
				require.Equal(`{}`, cresp.Text)
			})
		}
	})

	th.Run("user-bot-namespaces", func(th *Helper) {
		userData := model.StringInterface{
			"key1": "uservalue",
		}
		botData := model.StringInterface{
			"key2": "botvalue",
		}
		th.Cleanup(func() {
			_ = kvCall(th, "/delete", true, "", nil)
			_ = kvCall(th, "/delete", false, "", nil)
		})
		require := require.New(th)

		// Check that neither user's nor bot's test keys exist.
		cresp := kvCall(th, "/get", true, "", nil)
		require.Equal(`{}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "", nil)
		require.Equal(`{}`, cresp.Text)

		// Set and check both keys.
		cresp = kvCall(th, "/set", true, "", botData)
		require.Equal("true", cresp.Text)
		cresp = kvCall(th, "/set", false, "", userData)
		require.Equal("true", cresp.Text)
		cresp = kvCall(th, "/get", true, "", nil)
		require.Equal(`{"key2":"botvalue"}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "", nil)
		require.Equal(`{"key1":"uservalue"}`, cresp.Text)

		// Delete the user's, check that the bot's still there.
		cresp = kvCall(th, "/delete", false, "", nil)
		require.Equal(`deleted`, cresp.Text)
		cresp = kvCall(th, "/get", false, "", nil)
		require.Equal(`{}`, cresp.Text)
		cresp = kvCall(th, "/get", true, "", nil)
		require.Equal(`{"key2":"botvalue"}`, cresp.Text)

		// Create the user's key again, then delete bot's and re-test.
		cresp = kvCall(th, "/set", false, "", userData)
		require.Equal("true", cresp.Text)
		cresp = kvCall(th, "/delete", true, "", nil)
		require.Equal(`deleted`, cresp.Text)
		cresp = kvCall(th, "/get", true, "", nil)
		require.Equal(`{}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "", nil)
		require.Equal(`{"key1":"uservalue"}`, cresp.Text)
	})

	// TODO: Add a test for namespacing 2 separate users
}
