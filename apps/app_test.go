// +build !e2e

package apps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppConfigMap(t *testing.T) {
	m := &Manifest{
		AppID:                "app-id",
		DisplayName:          "display-name",
		Description:          "description",
		HomepageURL:          "homepage-url",
		HTTPRootURL:          "root_url",
		RequestedPermissions: Permissions{PermissionActAsUser, PermissionActAsBot},
		RequestedLocations:   Locations{LocationChannelHeader, LocationCommand},
	}

	a1 := &App{
		Manifest:           m,
		Secret:             "1234",
		OAuth2ClientID:     "id",
		OAuth2ClientSecret: "4321",
		OAuth2TrustedApp:   true,
		BotUserID:          "bot-user-id",
		BotUsername:        "bot_username",
		BotAccessToken:     "bot_access_token",
		GrantedPermissions: Permissions{PermissionActAsUser, PermissionActAsBot},
		GrantedLocations:   Locations{LocationChannelHeader, LocationCommand},
	}

	t.Run("App", func(t *testing.T) {
		map1 := a1.ConfigMap()
		// require.EqualValues(t, nil, map1)
		a2 := AppFromConfigMap(map1)
		require.EqualValues(t, a1, a2)

		map2 := a2.ConfigMap()
		require.EqualValues(t, map1, map2)
	})
}
