// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestValidateAppID(t *testing.T) {
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
			err := apps.AppID(id).Validate()
			if valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateAppVersion(t *testing.T) {
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
			err := apps.AppVersion(id).Validate()
			if valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		Manifest      apps.Manifest
		ExpectedError bool
	}{
		"empty manifest": {
			Manifest:      apps.Manifest{},
			ExpectedError: true,
		},
		"no deploy type": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"DisplayName empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: "https://example.org/root",
					},
				},
			},
			ExpectedError: true,
		},
		"HomepageURL empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: "https://example.org/root",
					},
				},
			},
			ExpectedError: true,
		},
		"HTTP RootURL empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{},
				},
			},
			ExpectedError: true,
		},
		"minimal valid HTTP app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: "https://example.org/root",
					},
				},
			},
			ExpectedError: false,
		},
		"invalid Icon": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: "https://example.org/root",
					},
				},
				Icon: "../..",
			},
			ExpectedError: true,
		},
		"invalid HomepageURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: ":invalid",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: "https://example.org/root",
					},
				},
			},
			ExpectedError: true,
		},
		"invalid HTTPRootURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org/root",
				Deploy: apps.Deploy{
					HTTP: &apps.HTTP{
						RootURL: ":invalid",
					},
				},
			},
			ExpectedError: true,
		},
		"no lambda for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{},
				},
			},
			ExpectedError: true,
		},
		"missing path for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{
						Functions: []apps.AWSLambdaFunction{{
							Name:    "go-funcion",
							Handler: "hello-lambda",
							Runtime: "go1.x",
						}},
					},
				},
			},
			ExpectedError: true,
		},
		"missing name for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{
						Functions: []apps.AWSLambdaFunction{{
							Path:    "/",
							Handler: "hello-lambda",
							Runtime: "go1.x",
						}},
					},
				},
			},
			ExpectedError: true,
		},
		"missing handler for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{
						Functions: []apps.AWSLambdaFunction{{
							Path:    "/",
							Name:    "go-funcion",
							Runtime: "go1.x",
						}},
					},
				},
			},
			ExpectedError: true,
		},
		"missing runtime for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{
						Functions: []apps.AWSLambdaFunction{{
							Path:    "/",
							Name:    "go-funcion",
							Handler: "hello-lambda",
						}},
					},
				},
			},
			ExpectedError: true,
		},
		"minimal valid AWS app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				DisplayName: "some display name",
				HomepageURL: "https://example.org",
				Deploy: apps.Deploy{
					AWSLambda: &apps.AWSLambda{
						Functions: []apps.AWSLambdaFunction{{
							Path:    "/",
							Name:    "go-funcion",
							Handler: "hello-lambda",
							Runtime: "go1.x",
						}},
					},
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
