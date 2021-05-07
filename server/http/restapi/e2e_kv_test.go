// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/api4"

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
		outSet, resp := th.BotClientPP.KVSet(id, prefix, in)
		api4.CheckOKStatus(t, resp)
		outSetMap, ok := outSet.(map[string]interface{})
		require.True(t, ok)
		require.Nil(t, resp.Error)
		require.Equal(t, outSetMap["changed"], true)

		// get
		var outGet map[string]interface{}
		resp = th.BotClientPP.KVGet(id, prefix, &outGet)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
		require.True(t, ok)
		require.Equal(t, outGet["test_bool"], true)
		require.Equal(t, outGet["test_string"], "test")

		// delete
		_, resp = th.BotClientPP.KVDelete(id, prefix)
		api4.CheckNoError(t, resp)
	})
}
