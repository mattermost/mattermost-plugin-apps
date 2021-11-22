//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/stretchr/testify/require"
)

func TestKVE2E(t *testing.T) {
	th := Setup(t)
	th.SetupPP()

	app := th.SetupApp(apps.Manifest{
		AppID:       apps.AppID("some_app_id"),
		DisplayName: "Some Display Name",
	})

	t.Run("test KV API", func(t *testing.T) {
		id := "testId"
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		// set
		changed, resp, err := app.AsBot.KVSet(id, prefix, in)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		require.True(t, changed)

		// get
		var outGet map[string]interface{}
		resp, err = app.AsBot.KVGet(id, prefix, &outGet)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		require.Equal(t, outGet["test_bool"], true)
		require.Equal(t, outGet["test_string"], "test")

		// delete
		_, err = app.AsBot.KVDelete(id, prefix)
		require.NoError(t, err)
	})
}
