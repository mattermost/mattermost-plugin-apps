package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/stretchr/testify/require"
)

func TestSubscribe(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	th.ServerTestHelper.LoginSystemAdmin()

	subscription := &apps.Subscription{
		AppID:     "test-apiId",
		Subject:   "test-subject",
		ChannelID: th.ServerTestHelper.BasicChannel.Id,
		TeamID:    th.ServerTestHelper.BasicTeam.Id,
	}

	th.TestForSystemAdmin(t, func(t *testing.T, client *mmclient.ClientPP) {
		_, resp := client.Subscribe(subscription)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
	})
}

func TestUnsubscribe(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	th.ServerTestHelper.LoginSystemAdmin()

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
}
