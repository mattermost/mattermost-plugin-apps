//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribeE2E(t *testing.T) {
	th := Setup(t)
	SetupPP(th, t)
	defer th.TearDown()

	t.Run("Unauthenticated requests are rejected", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		client := th.CreateClientPP()

		resp, err := client.Subscribe(subscription)
		require.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)

		subs, resp, err := client.GetSubscriptions()
		require.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
		require.Nil(t, subs)

		resp, err = client.Unsubscribe(subscription)
		require.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
	})

	t.Run("Users can delete there own subscriptions", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}

		th.TestForUserAndSystemAdmin(t, func(t *testing.T, client *appclient.ClientPP) {
			// Subscribe
			resp, err := client.Subscribe(subscription)
			require.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			// List subscriptions
			subs, resp, err := client.GetSubscriptions()
			require.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Len(t, subs, 1)

			// Unsubscribe
			resp, err = client.Unsubscribe(subscription)
			require.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			// List subscriptions
			subs, resp, err = client.GetSubscriptions()
			require.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Len(t, subs, 0)
		})
	})

	t.Run("Users can't delete other users subscriptions", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		resp, err := th.SystemAdminClientPP.Subscribe(subscription)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		resp, err = th.ClientPP.Unsubscribe(subscription)
		require.Error(t, err)
		api4.CheckNotFoundStatus(t, resp)
	})

	t.Run("Users can't see other users subscriptions", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		resp, err := th.SystemAdminClientPP.Subscribe(subscription)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		resp, err = th.ClientPP.Subscribe(subscription)
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		subs, resp, err := th.ClientPP.GetSubscriptions()
		require.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.Len(t, subs, 1)
	})

	t.Run("Bad request for missing subject", func(t *testing.T) {
		subscription := &apps.Subscription{
			AppID:     "test-apiId",
			Subject:   "",
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		resp, err := th.SystemAdminClientPP.Subscribe(subscription)
		require.Error(t, err)
		api4.CheckBadRequestStatus(t, resp)
	})
}
