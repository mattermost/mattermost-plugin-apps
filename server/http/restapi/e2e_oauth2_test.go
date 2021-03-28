// +build e2e

package restapi

import (
	"testing"
)

func TestRemoteOAuth2E2E(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	t.Run("test Remote OAuth2 API", func(t *testing.T) {
		// id := "testId"
		// prefix := "prefix-test"
		// in := map[string]interface{}{}
		// in["test_bool"] = true
		// in["test_string"] = "test"

		// // set
		// outSet, resp := th.BotClientPP.KVSet(id, prefix, in)
		// api4.CheckOKStatus(t, resp)
		// outSetMap, ok := outSet.(map[string]interface{})
		// require.True(t, ok)
		// require.Nil(t, resp.Error)
		// require.Equal(t, outSetMap["changed"], true)

		// // get
		// outGet, resp := th.BotClientPP.KVGet(id, prefix)
		// api4.CheckOKStatus(t, resp)
		// require.Nil(t, resp.Error)
		// outGetMap, ok := outGet.(map[string]interface{})
		// require.True(t, ok)
		// require.Equal(t, outGetMap["test_bool"], true)
		// require.Equal(t, outGetMap["test_string"], "test")

		// // delete
		// _, resp = th.BotClientPP.KVDelete(id, prefix)
		// api4.CheckNoError(t, resp)
	})
}
