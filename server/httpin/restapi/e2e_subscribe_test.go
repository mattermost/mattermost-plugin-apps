// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"

	"github.com/stretchr/testify/require"
)

func TestSubscribeE2E(t *testing.T) {
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

		th.TestForSystemAdmin(t, func(t *testing.T, client *appclient.ClientPP) {
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
}
