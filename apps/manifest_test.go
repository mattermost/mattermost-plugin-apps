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
		"missing app type": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"HomepageURL empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HTTPRootURL: "https://example.org/root",
			},
			ExpectedError: true,
		},
		"HTTPRootURL empty": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"minimal valid HTTP app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HomepageURL: "https://example.org",
				HTTPRootURL: "https://example.org/root",
			},
			ExpectedError: false,
		},
		"invalid Icon": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HomepageURL: "https://example.org",
				HTTPRootURL: "https://example.org/root",
				Icon:        "../..",
			},
			ExpectedError: true,
		},
		"invalid HomepageURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HomepageURL: ":invalid",
				HTTPRootURL: "https://example.org/root",
			},
			ExpectedError: true,
		},
		"invalid HTTPRootURL": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeHTTP,
				HomepageURL: "https://example.org/root",
				HTTPRootURL: ":invalid",
			},
			ExpectedError: true,
		},
		"no lambda for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
			},
			ExpectedError: true,
		},
		"missing path for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
				AWSLambda: []apps.AWSLambda{{
					Name:    "go-funcion",
					Handler: "hello-lambda",
					Runtime: "go1.x",
				}},
			},
			ExpectedError: true,
		},
		"missing name for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
				AWSLambda: []apps.AWSLambda{{
					Path:    "/",
					Handler: "hello-lambda",
					Runtime: "go1.x",
				}},
			},
			ExpectedError: true,
		},
		"missing handler for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
				AWSLambda: []apps.AWSLambda{{
					Path:    "/",
					Name:    "go-funcion",
					Runtime: "go1.x",
				}},
			},
			ExpectedError: true,
		},
		"missing runtime for AWS app": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
				AWSLambda: []apps.AWSLambda{{
					Path:    "/",
					Name:    "go-funcion",
					Handler: "hello-lambda",
				}},
			},
			ExpectedError: true,
		},
		"minimal valid AWS app example manifest": {
			Manifest: apps.Manifest{
				AppID:       "abc",
				AppType:     apps.AppTypeAWSLambda,
				HomepageURL: "https://example.org",
				AWSLambda: []apps.AWSLambda{{
					Path:    "/",
					Name:    "go-funcion",
					Handler: "hello-lambda",
					Runtime: "go1.x",
				}},
			},
			ExpectedError: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := test.Manifest.IsValid()
			if test.ExpectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
