//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestAppInstallE2E(t *testing.T) {
	th := Setup(t)
	th.SetupPP()

	th.SetupApp(apps.Manifest{
		AppID:       apps.AppID("some_app_id"),
		DisplayName: "Some Display Name",
	})
}

func TestSubscribeE2E(t *testing.T) {
	th := Setup(t)
	th.SetupPP()

	app := th.SetupApp(apps.Manifest{
		AppID:       apps.AppID("some_app_id"),
		DisplayName: "Some Display Name",
	})

	t.Run("Unauthenticated requests are rejected", func(t *testing.T) {
		subscription := &apps.Subscription{
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		client := th.CreateClientPP()

		resp, err := client.Subscribe(subscription)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)

		subs, resp, err := client.GetSubscriptions()
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
		assert.Nil(t, subs)

		resp, err = client.Unsubscribe(subscription)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
	})

	t.Run("Users can delete there own subscriptions", func(t *testing.T) {
		subscription := apps.Subscription{
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}

		app.TestForTwoUsersAndBot(func(t *testing.T, client *TestClientPP) {
			// Subscribe

			resp, err := client.Subscribe(&subscription)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			// List subscriptions
			subs, resp, err := client.GetSubscriptions()
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Len(t, subs, 1)

			expectedSub := subscription
			expectedSub.AppID = app.Manifest.AppID
			expectedSub.UserID = client.UserID
			assert.Equal(t, expectedSub, subs[0])

			// Unsubscribe
			resp, err = client.Unsubscribe(&subscription)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			// List subscriptions
			subs, resp, err = client.GetSubscriptions()
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Len(t, subs, 0)
		})
	})

	t.Run("Users can't delete other users subscriptions", func(t *testing.T) {
		subscription := &apps.Subscription{
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}

		resp, err := app.AsUser.Subscribe(subscription)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		resp, err = app.AsUser2.Unsubscribe(subscription)
		assert.Error(t, err)
		api4.CheckNotFoundStatus(t, resp)

		subs, resp, err := app.AsUser.GetSubscriptions()
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.Len(t, subs, 1)
	})

	t.Run("Users can't see other users subscriptions", func(t *testing.T) {
		subscription := &apps.Subscription{
			Subject:   apps.SubjectUserJoinedChannel,
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		resp, err := app.AsUser.Subscribe(subscription)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		resp, err = app.AsUser2.Subscribe(subscription)
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)

		subs, resp, err := app.AsUser2.GetSubscriptions()
		assert.NoError(t, err)
		api4.CheckOKStatus(t, resp)
		assert.Len(t, subs, 1)
	})

	t.Run("Bad request for missing subject", func(t *testing.T) {
		subscription := &apps.Subscription{
			Subject:   "",
			ChannelID: th.ServerTestHelper.BasicChannel.Id,
			TeamID:    th.ServerTestHelper.BasicTeam.Id,
			Call: apps.Call{
				Path: "/some/path",
			},
		}
		resp, err := app.AsUser.Subscribe(subscription)
		assert.Error(t, err)
		api4.CheckBadRequestStatus(t, resp)
	})
}
