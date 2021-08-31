// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestIsValid(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		Manifest      Manifest
		ExpectedError bool
	}{
		"empty manifest": {
			Manifest:      Manifest{},
			ExpectedError: true,
		},
		"no app types": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"HomepageURL empty": {
			Manifest: Manifest{
				AppID: "abc",
				HTTP: &HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: true,
		},
		"HTTP RootURL empty": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP:        &HTTP{},
			},
			ExpectedError: true,
		},
		"minimal valid HTTP app example manifest": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP: &HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: false,
		},
		"invalid Icon": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP: &HTTP{
					RootURL: "https://example.org/root",
				},
				Icon: "../..",
			},
			ExpectedError: true,
		},
		"invalid HomepageURL": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: ":invalid",
				HTTP: &HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: true,
		},
		"invalid HTTPRootURL": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org/root",
				HTTP: &HTTP{
					RootURL: ":invalid",
				},
			},
			ExpectedError: true,
		},
		"no lambda for AWS app": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda:   &AWSLambda{},
			},
			ExpectedError: true,
		},
		"missing path for AWS app": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &AWSLambda{
					Functions: []AWSLambdaFunction{{
						Name:    "go-funcion",
						Handler: "hello-lambda",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing name for AWS app": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &AWSLambda{
					Functions: []AWSLambdaFunction{{
						Path:    "/",
						Handler: "hello-lambda",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing handler for AWS app": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &AWSLambda{
					Functions: []AWSLambdaFunction{{
						Path:    "/",
						Name:    "go-funcion",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing runtime for AWS app": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &AWSLambda{
					Functions: []AWSLambdaFunction{{
						Path:    "/",
						Name:    "go-funcion",
						Handler: "hello-lambda",
					}},
				},
			},
			ExpectedError: true,
		},
		"minimal valid AWS app example manifest": {
			Manifest: Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &AWSLambda{
					Functions: []AWSLambdaFunction{{
						Path:    "/",
						Name:    "go-funcion",
						Handler: "hello-lambda",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := test.Manifest.Validate()

			if test.ExpectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManifestUnmarshalJSON(t *testing.T) {
	hello := Manifest{
		AppID:                "hello-world",
		DisplayName:          "Hello, world!",
		Icon:                 "icon.png",
		HomepageURL:          "http://localhost:8080",
		RequestedPermissions: Permissions{PermissionActAsBot},
		RequestedLocations:   Locations{LocationChannelHeader, LocationCommand},
	}

	helloHTTP := hello
	helloHTTP.HTTP = &HTTP{
		RootURL: "http://localhost:8080",
	}
	helloHTTP7 := helloHTTP
	helloHTTP7.v7AppType = string(DeployHTTP)

	helloPlugin := hello
	helloPlugin.Plugin = &Plugin{
		PluginID: "com.mattermost.hello-world",
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"app_type": "http",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
					"root_url": "http://localhost:8080",
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
					"http":{
						"root_url": "http://localhost:8080"
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"app_type": "aws_lambda",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"app_type": "plugin",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
					"plugin_id": "com.mattermost.hello-world",
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
					"app_id": "hello-world",
					"display_name": "Hello, world!",
					"icon": "icon.png",
					"homepage_url":"http://localhost:8080",
					"plugin": {
						"plugin_id": "com.mattermost.hello-world"
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
