//go:build e2e
// +build e2e

package restapi

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v6/api4"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestOAuth2UserE2E(t *testing.T) {
	th := Setup(t)
	th.SetupPP()

	app := th.SetupApp(apps.Manifest{
		AppID:       apps.AppID("some_app_id"),
		DisplayName: "Some Display Name",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsBot,
			apps.PermissionActAsUser,
			apps.PermissionRemoteOAuth2,
		},
	})

	t.Run("Unauthenticated requests are rejected", func(t *testing.T) {
		in := map[string]interface{}{
			"test_bool":   true,
			"test_string": "test",
		}

		client := th.CreateClientPP()
		resp, err := client.StoreOAuth2User(in)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)

		var out map[string]interface{}
		resp, err = client.GetOAuth2User(&out)
		assert.Error(t, err)
		api4.CheckUnauthorizedStatus(t, resp)
	})

	t.Run("Users can set and get there own OAuth2User", func(t *testing.T) {
		app.TestForUser(func(t *testing.T, client *TestClientPP) {
			in := map[string]interface{}{
				"test_bool":   true,
				"test_string": "test",
			}
			resp, err := client.StoreOAuth2User(in)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)

			var out map[string]interface{}
			resp, err = client.GetOAuth2User(&out)
			assert.NoError(t, err)
			api4.CheckOKStatus(t, resp)
			assert.Equal(t, in, out)
		})
	})
}
