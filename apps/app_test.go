// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppIDIsValid(t *testing.T) {
	t.Parallel()

	for id, valid := range map[string]bool{
		"":                                  false,
		"a":                                 false,
		"ab":                                false,
		"abc":                               true,
		"abcdefghijklmnopqrstuvwxyzabcdef":  true,
		"abcdefghijklmnopqrstuvwxyzabcdefg": false,
		"../path":                           false,
		"/etc/passwd":                       false,
		"com.mattermost.app-0.9":            true,
		"CAPS-ARE-FINE":                     true,
		"....DOTS.ALSO.......":              true,
		"----SLASHES-ALSO----":              true,
		"___AND_UNDERSCORES____":            true,
	} {
		t.Run(id, func(t *testing.T) {
			err := AppID(id).Validate()
			if valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAppVersionIsValid(t *testing.T) {
	t.Parallel()

	for id, valid := range map[string]bool{
		"":            true,
		"v1.0.0":      true,
		"1.0.0":       true,
		"v1.0.0-rc1":  true,
		"1.0.0-rc1":   true,
		"CAPS-OK":     true,
		".DOTS.":      true,
		"-SLASHES-":   true,
		"_OK_":        true,
		"v00_00_0000": false,
		"/":           false,
	} {
		t.Run(id, func(t *testing.T) {
			err := AppVersion(id).Validate()
			if valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAppUnmarshalJSON(t *testing.T) {
	hello := App{
		DeployType: DeployHTTP,
		Manifest: Manifest{
			AppID:                "hello-world",
			DisplayName:          "Hello, world!",
			Icon:                 "icon.png",
			HomepageURL:          "http://localhost:8080",
			RequestedPermissions: Permissions{PermissionActAsBot},
			RequestedLocations:   Locations{LocationChannelHeader, LocationCommand},
			HTTP: &HTTP{
				RootURL: "http://localhost:8080",
			},
			v7AppType: "http",
		},
		BotUserID:          "egfof4cejjr3bmcxbepexukuaa",
		BotUsername:        "hello-world",
		BotAccessToken:     "fdpudxyqt7n3jyyor3mc34unty",
		BotAccessTokenID:   "xeaq7yqns3n6dbyix53x3j7m1y",
		GrantedPermissions: Permissions{PermissionActAsBot},
		GrantedLocations:   Locations{LocationChannelHeader, LocationCommand},
	}

	for name, test := range map[string]struct {
		In            string
		Expected      App
		ExpectedError string
	}{
		"happy v7": {
			In: `{
				"app_id":"hello-world",
				"app_type":"http",
				"version":"",
				"homepage_url":"http://localhost:8080",
				"display_name":"Hello, world!",
				"icon":"icon.png",
				"requested_permissions":["act_as_bot"],
				"requested_locations":["/channel_header","/command"],
				"root_url":"http://localhost:8080",
				"bot_user_id":"egfof4cejjr3bmcxbepexukuaa",
				"bot_username":"hello-world",
				"bot_access_token":"fdpudxyqt7n3jyyor3mc34unty",
				"bot_access_token_id":"xeaq7yqns3n6dbyix53x3j7m1y",
				"mattermost_oauth2":{},
				"remote_oauth2":{},
				"granted_permissions":["act_as_bot"],
				"granted_locations":["/channel_header","/command"]
			}`,
			Expected: hello,
		},
	} {
		t.Run(name, func(t *testing.T) {
			app, err := DecodeCompatibleApp([]byte(test.In))
			if test.ExpectedError != "" {
				require.Error(t, err)
				require.Equal(t, test.ExpectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.EqualValues(t, test.Expected, *app)
			}
		})
	}
}
