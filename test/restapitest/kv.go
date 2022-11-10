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
			Version:     "v1.2.0",
			DisplayName: "tests access to the KV store",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/test/restapitest",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
	)

	params := func(creq goapp.CallRequest) (client *appclient.Client, prefix, key string, value interface{}) {
		prefix, _ = creq.StringValue("prefix")
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
			return appclient.AsBot(creq.Context), prefix, key, value
		}
		require.NotEmpty(t, creq.Context.ActingUserAccessToken)
		return appclient.AsActingUser(creq.Context), prefix, key, value
	}

	app.HandleCall("/get",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, prefix, key, _ := params(creq)
			v := map[string]interface{}{}
			err := client.KVGet(prefix, key, &v)
			require.NoError(t, err)
			return apps.NewTextResponse(utils.ToJSON(v))
		})

	app.HandleCall("/set",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, prefix, key, value := params(creq)
			changed, err := client.KVSet(prefix, key, value)
			require.NoError(t, err)
			return apps.NewTextResponse("%v", changed)
		})

	app.HandleCall("/delete",
		func(creq goapp.CallRequest) apps.CallResponse {
			client, prefix, key, _ := params(creq)
			err := client.KVDelete(prefix, key)
			require.NoError(t, err)
			return apps.NewTextResponse("deleted")
		})

	return app
}

func kvCall(th *Helper, path string, asBot bool, prefix, key string, value interface{}) apps.CallResponse {
	creq := apps.CallRequest{
		Call: *apps.NewCall(path),
		Values: model.StringInterface{
			"as_bot": asBot,
		},
	}
	if !asBot {
		creq.Call.Expand = &apps.Expand{
			ActingUser:            apps.ExpandSummary.Required(),
			ActingUserAccessToken: apps.ExpandAll,
		}
	}
	if prefix != "" {
		creq.Values["prefix"] = prefix
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
	th.InstallAppWithCleanup(newKVApp(th.T))

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
				// Check that does not yet exist, should comes back empty.
				cresp := kvCall(th, "/get", asBot, "pp", "", nil)
				require.Equal(th, `{}`, cresp.Text)

				// Set.
				cresp = kvCall(th, "/set", asBot, "pp", "", nil)
				require.Equal(th, "true", cresp.Text)

				// Check.
				cresp = kvCall(th, "/get", asBot, "pp", "", nil)
				require.Equal(th, `{"test-name":"test-data"}`, cresp.Text)

				// Delete.
				cresp = kvCall(th, "/delete", asBot, "pp", "", nil)
				require.Equal(th, `deleted`, cresp.Text)

				// Check again - deleted comes back as empty.
				cresp = kvCall(th, "/get", asBot, "pp", "", nil)
				require.Equal(th, `{}`, cresp.Text)
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
			_ = kvCall(th, "/delete", true, "pp", "", nil)
			_ = kvCall(th, "/delete", false, "pp", "", nil)
			th.Logf("deleted test KV keys")
		})

		// Check that neither user's nor bot's test keys exist.
		cresp := kvCall(th, "/get", true, "pp", "", nil)
		require.Equal(th, `{}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "pp", "", nil)
		require.Equal(th, `{}`, cresp.Text)

		// Set and check both keys.
		cresp = kvCall(th, "/set", true, "pp", "", botData)
		require.Equal(th, "true", cresp.Text)
		cresp = kvCall(th, "/set", false, "pp", "", userData)
		require.Equal(th, "true", cresp.Text)
		cresp = kvCall(th, "/get", true, "pp", "", nil)
		require.Equal(th, `{"key2":"botvalue"}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "pp", "", nil)
		require.Equal(th, `{"key1":"uservalue"}`, cresp.Text)

		// Delete the user's, check that the bot's still there.
		cresp = kvCall(th, "/delete", false, "pp", "", nil)
		require.Equal(th, `deleted`, cresp.Text)
		cresp = kvCall(th, "/get", false, "pp", "", nil)
		require.Equal(th, `{}`, cresp.Text)
		cresp = kvCall(th, "/get", true, "pp", "", nil)
		require.Equal(th, `{"key2":"botvalue"}`, cresp.Text)

		// Create the user's key again, then delete bot's and re-test.
		cresp = kvCall(th, "/set", false, "pp", "", userData)
		require.Equal(th, "true", cresp.Text)
		cresp = kvCall(th, "/delete", true, "pp", "", nil)
		require.Equal(th, `deleted`, cresp.Text)
		cresp = kvCall(th, "/get", true, "pp", "", nil)
		require.Equal(th, `{}`, cresp.Text)
		cresp = kvCall(th, "/get", false, "pp", "", nil)
		require.Equal(th, `{"key1":"uservalue"}`, cresp.Text)
	})

	th.Run("prefixes", func(th *Helper) {
		noPrefixData := model.StringInterface{
			"key1": "no prefix",
		}
		p1Data := model.StringInterface{
			"key2": "p1",
		}
		p2Data := model.StringInterface{
			"key3": "p2",
		}
		th.Cleanup(func() {
			_ = kvCall(th, "/delete", false, "", "testkey", nil)
			_ = kvCall(th, "/delete", false, "p1", "testkey", nil)
			_ = kvCall(th, "/delete", false, "p2", "testkey", nil)
			th.Logf("deleted test KV keys")
		})

		// Set and check both keys.
		setAndVerify := func(prefix, expected string, data model.StringInterface) {
			cresp := kvCall(th, "/set", false, prefix, "testkey", data)
			require.Equal(th, "true", cresp.Text)
			cresp = kvCall(th, "/get", false, prefix, "testkey", nil)
			require.Equal(th, expected, cresp.Text)
		}
		setAndVerify("", `{"key1":"no prefix"}`, noPrefixData)
		setAndVerify("p1", `{"key2":"p1"}`, p1Data)
		setAndVerify("p2", `{"key3":"p2"}`, p2Data)
	})

	// TODO: Add a test for namespacing 2 separate users
}
