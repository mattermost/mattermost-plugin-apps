package apps_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestManifestIsValid(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		Manifest      apps.Manifest
		ExpectedError bool
	}{
		"empty manifest": {
			Manifest:      apps.Manifest{},
			ExpectedError: true,
		},
		"no app types": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"HomepageURL empty": {
			Manifest: apps.Manifest{
				AppID: "abc",
				HTTP: &apps.HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: true,
		},
		"HTTP RootURL empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP:        &apps.HTTP{},
			},
			ExpectedError: true,
		},
		"minimal valid HTTP app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP: &apps.HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: false,
		},
		"invalid Icon": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				HTTP: &apps.HTTP{
					RootURL: "https://example.org/root",
				},
				Icon: "../..",
			},
			ExpectedError: true,
		},
		"invalid HomepageURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: ":invalid",
				HTTP: &apps.HTTP{
					RootURL: "https://example.org/root",
				},
			},
			ExpectedError: true,
		},
		"invalid HTTPRootURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org/root",
				HTTP: &apps.HTTP{
					RootURL: ":invalid",
				},
			},
			ExpectedError: true,
		},
		"no lambda for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda:   &apps.AWSLambda{},
			},
			ExpectedError: true,
		},
		"missing path for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &apps.AWSLambda{
					Functions: []apps.AWSLambdaFunction{{
						Name:    "go-funcion",
						Handler: "hello-lambda",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing name for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &apps.AWSLambda{
					Functions: []apps.AWSLambdaFunction{{
						Path:    "/",
						Handler: "hello-lambda",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing handler for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &apps.AWSLambda{
					Functions: []apps.AWSLambdaFunction{{
						Path:    "/",
						Name:    "go-funcion",
						Runtime: "go1.x",
					}},
				},
			},
			ExpectedError: true,
		},
		"missing runtime for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &apps.AWSLambda{
					Functions: []apps.AWSLambdaFunction{{
						Path:    "/",
						Name:    "go-funcion",
						Handler: "hello-lambda",
					}},
				},
			},
			ExpectedError: true,
		},
		"minimal valid AWS app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
				AWSLambda: &apps.AWSLambda{
					Functions: []apps.AWSLambdaFunction{{
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
