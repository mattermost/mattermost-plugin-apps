//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/stretchr/testify/assert"
)

func TestKVE2E(t *testing.T) {
	th := Setup(t)
	th.SetupPP()

	app := th.SetupApp(apps.Manifest{
		AppID:       apps.AppID("some_app_id"),
		DisplayName: "Some Display Name",
	})

	t.Run("Unauthenticated requests are rejected", func(t *testing.T) {
		id := model.NewId()
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		client := th.CreateClientPP()

		changed, resp, err := client.KVSet(prefix, id, in)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
		assert.False(t, changed)

		var out map[string]interface{}
		resp, err = client.KVGet(prefix, id, &out)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
		assert.False(t, changed)

		resp, err = client.KVDelete(prefix, id)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
	})

	t.Run("Users create, get and delete KV entries", func(t *testing.T) {
		id := model.NewId()
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		app.TestForTwoUsersAndBot(func(t *testing.T, client *TestClientPP) {
			changed, resp, err := client.KVSet(prefix, id, in)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.True(t, changed)

			var out map[string]interface{}
			resp, err = client.KVGet(prefix, id, &out)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Equal(t, out["test_bool"], true)
			assert.Equal(t, out["test_string"], "test")

			resp, err = client.KVDelete(prefix, id)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			out = nil
			resp, err = client.KVGet(prefix, id, &out)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Nil(t, out)
		})
	})

	t.Run("Users can't delete other users KV entries", func(t *testing.T) {
		id := model.NewId()
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		changed, resp, err := app.AsUser.KVSet(prefix, id, in)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.True(t, changed)

		// Deleting something that isn't there still returns a 200
		resp, err = app.AsUser2.KVDelete(prefix, id)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		var out map[string]interface{}
		resp, err = app.AsUser.KVGet(prefix, id, &out)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.Equal(t, out["test_bool"], true)
		assert.Equal(t, out["test_string"], "test")
	})

	t.Run("Users can't see other users KV entires", func(t *testing.T) {
		id := model.NewId()
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		changed, resp, err := app.AsUser.KVSet(prefix, id, in)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.True(t, changed)

		var out map[string]interface{}
		resp, err = app.AsUser2.KVGet(prefix, id, &out)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.Nil(t, out)
	})
}
