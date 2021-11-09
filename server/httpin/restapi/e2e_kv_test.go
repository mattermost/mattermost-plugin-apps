//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/stretchr/testify/require"
)

func TestKVE2E(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	t.Run("test KV API", func(t *testing.T) {
		id := "testId"
		prefix := "PT"
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		// set
		changed, resp, err := th.BotClientPP.KVSet(id, prefix, in)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		require.True(t, changed)

		// get
		var outGet map[string]interface{}
		resp, err = th.BotClientPP.KVGet(id, prefix, &outGet)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		require.Equal(t, outGet["test_bool"], true)
		require.Equal(t, outGet["test_string"], "test")

		// delete
		_, err = th.BotClientPP.KVDelete(id, prefix)
		require.NoError(t, err)
	})
}
