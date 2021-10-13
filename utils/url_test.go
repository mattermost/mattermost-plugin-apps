package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestIsValidHttpUrl(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		URL           string
		ExpectedError bool
	}{

		"empty url": {
			"",
			true,
		},
		"bad url": {
			"bad url",
			true,
		},
		"relative url": {
			"/api/test",
			true,
		},
		"relative url ending with slash": {
			"/some/url/",
			true,
		},
		"url with invalid scheme": {
			"htp://mattermost.com",
			true,
		},
		"url with just http": {
			"http://",
			true,
		},
		"url with just https": {
			"https://",
			true,
		},
		"url with extra slashes": {
			"https:///mattermost.com",
			true,
		},
		"correct url with http scheme": {
			"http://mattemost.com",
			false,
		},
		"correct url with https scheme": {
			"https://mattermost.com/api/test",
			false,
		},
		"correct url with port": {
			"https://localhost:1111/test",
			false,
		},
		"correct url without scheme": {
			"mattermost.com/some/url/",
			true,
		},
		"correct url with extra slashes": {
			"https://mattermost.com/some//url",
			false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := utils.IsValidHTTPURL(test.URL)

			if test.ExpectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
