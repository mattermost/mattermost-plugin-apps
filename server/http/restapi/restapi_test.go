// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"

	"github.com/stretchr/testify/require"
)

func TestPPAPI(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	t.Run("test Subscribe API", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   "test-subject",
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
		}

		th.TestForSystemAdmin(t, func(t *testing.T, client *mmclient.ClientPP) {
			// subscribe
			_, resp := client.Subscribe(subscription)
			api4.CheckOKStatus(t, resp)
			require.Nil(t, resp.Error)

			// unsubscribe
			_, resp = client.Unsubscribe(subscription)
			api4.CheckOKStatus(t, resp)
			require.Nil(t, resp.Error)
		})
	})

	t.Run("test KV API", func(t *testing.T) {
		id := "testId"
		prefix := "prefix-test"
		in := map[string]interface{}{}
		in["test_bool"] = true
		in["test_string"] = "test"

		// set
		outSet, resp := th.BotClientPP.KVSet(id, prefix, in)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, outSet["changed"], true)

		// get
		outGet, resp := th.BotClientPP.KVGet(id, prefix)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
		require.Equal(t, outGet["test_bool"], true)
		require.Equal(t, outGet["test_string"], "test")

		// delete
		_, resp = th.BotClientPP.KVDelete(id, prefix)
		api4.CheckNoError(t, resp)
	})
}
