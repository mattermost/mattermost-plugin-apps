// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManifestUnmarshalJSON(t *testing.T) {
	hello := Manifest{
		AppID:                "hello-test",
		DisplayName:          "Hello, test!",
		Icon:                 "icon.png",
		HomepageURL:          "http://localhost:1111",
		RequestedPermissions: Permissions{PermissionActAsBot},
		RequestedLocations:   Locations{LocationChannelHeader, LocationCommand},
	}

	helloHTTP := hello
	helloHTTP.HTTP = &HTTP{
		RootURL: "http://localhost:1111",
	}

	helloPlugin := hello
	helloPlugin.Plugin = &Plugin{
		PluginID: "com.mattermost.hello-test",
	}

	helloAWS := hello
	helloAWS.AWSLambda = &AWSLambda{
		Functions: []AWSLambdaFunction{
			{
				Path:    "/",
				Name:    "go-function",
				Handler: "hello-lambda",
				Runtime: "go1.x",
			},
		},
	}

	for name, test := range map[string]struct {
		In            string
		Expected      Manifest
		ExpectedError string
	}{
		"http": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"http":{
						"root_url": "http://localhost:1111"
					},
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloHTTP,
		},
		"aws": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"aws_lambda": {
						"functions": [
							{
								"path": "/",
								"name": "go-function",
								"handler": "hello-lambda",
								"runtime": "go1.x"
							}
						]
					},
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloAWS,
		},
		"plugin": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"plugin": {
						"plugin_id": "com.mattermost.hello-test"
					},
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloPlugin,
		},
	} {
		t.Run(name, func(t *testing.T) {
			m, err := DecodeCompatibleManifest([]byte(test.In))
			if test.ExpectedError != "" {
				require.Error(t, err)
				require.Equal(t, test.ExpectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.EqualValues(t, test.Expected, *m)
			}
		})
	}
}
