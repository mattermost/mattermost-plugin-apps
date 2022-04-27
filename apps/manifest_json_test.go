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
	helloHTTP7 := helloHTTP
	helloHTTP7.v7AppType = string(DeployHTTP)

	helloPlugin := hello
	helloPlugin.Plugin = &Plugin{
		PluginID: "com.mattermost.hello-test",
	}
	helloPlugin7 := helloPlugin
	helloPlugin7.v7AppType = string(DeployPlugin)

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
	helloAWS7 := helloAWS
	helloAWS7.v7AppType = string(DeployAWSLambda)

	for name, test := range map[string]struct {
		In            string
		Expected      Manifest
		ExpectedError string
	}{
		"v0.7 http": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"app_type": "http",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"root_url": "http://localhost:1111",
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloHTTP7,
		},
		"v0.8 http": {
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
		"v0.7 aws": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"app_type": "aws_lambda",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"aws_lambda": [
						{
							"path": "/",
							"name": "go-function",
							"handler": "hello-lambda",
							"runtime": "go1.x"
						}
					],
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloAWS7,
		},
		"v0.8 aws": {
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
		"v0.7 plugin": {
			In: `{
					"app_id": "hello-test",
					"display_name": "Hello, test!",
					"app_type": "plugin",
					"icon": "icon.png",
					"homepage_url":"http://localhost:1111",
					"plugin_id": "com.mattermost.hello-test",
					"requested_permissions": [
						"act_as_bot"
					],
					"requested_locations": [
						"/channel_header",
						"/command"
					]
				}`,
			Expected: helloPlugin7,
		},
		"v0.8 plugin": {
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
