package store

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestHashkey(t *testing.T) {
	for _, tc := range []struct {
		name               string
		globalPrefix       string
		appID              apps.AppID
		userID, prefix, id string
		expectedError      string
		expected           string
	}{
		{
			name:         "long",
			globalPrefix: ".X",
			appID:        "some_app_id",
			userID:       "userIDis26bytes12345678910",
			prefix:       "",
			id:           "the-app-chose-an-even-very-longer-id-----------------",
			expected:     ".Xsome_app_id                     userIDis26bytes12345678910  DONVG5fTTQ:B^jR&Hc+M",
		},

		{
			name:         "one byte prefix",
			globalPrefix: ".X",
			appID:        "some_app_id",
			userID:       "userIDis26bytes12345678910",
			prefix:       "A",
			id:           "id",
			expected:     ".Xsome_app_id                     userIDis26bytes12345678910A nEDMRpHe-;lXX_IkLt)+",
		},

		{
			name:         "two byte prefix",
			globalPrefix: ".X",
			appID:        "some_app_id",
			userID:       "userIDis26bytes12345678910",
			prefix:       "AB",
			id:           "id",
			expected:     ".Xsome_app_id                     userIDis26bytes12345678910ABnEDMRpHe-;lXX_IkLt)+",
		},
		{
			name:         "short",
			globalPrefix: ".X",
			appID:        "some_app_id",
			userID:       "userIDis26bytes12345678910",
			prefix:       "",
			id:           "0",
			expected:     ".Xsome_app_id                     userIDis26bytes12345678910  @ZG`[Gt?1\"TiTJ%@i@0n",
		},
		{
			name:          "error empty globalPrefix",
			globalPrefix:  "",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix does not start with a dot",
			globalPrefix:  "AB",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "AB" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix too short",
			globalPrefix:  ".",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "." is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix too long",
			globalPrefix:  ".TOOLONG",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix ".TOOLONG" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error userID too short",
			globalPrefix:  ".X",
			appID:         "some_app_id",
			userID:        "TOOSHORT",
			prefix:        "A",
			id:            "id",
			expectedError: `userID "TOOSHORT" must be exactly 26 ASCII characters`,
		},
		{
			name:          "error userID too long",
			globalPrefix:  ".X",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910TOOLONG",
			prefix:        "A",
			id:            "id",
			expectedError: `userID "userIDis26bytes12345678910TOOLONG" must be exactly 26 ASCII characters`,
		},
		{
			name:          "error id empty",
			globalPrefix:  ".X",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "A",
			id:            "",
			expectedError: `key must not be empty`,
		},
		{
			name:          "error prefix too long",
			globalPrefix:  ".X",
			appID:         "some_app_id",
			userID:        "userIDis26bytes12345678910",
			prefix:        "ABC",
			id:            "id",
			expectedError: `prefix "ABC" is longer than the limit of 2 ASCII characters`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, err := Hashkey(tc.globalPrefix, tc.appID, tc.userID, tc.prefix, tc.id)
			if tc.expectedError != "" {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, hashKeyLength, len(key))
				assert.Equal(t, tc.expected, key)

				gp, a, u, p, h, err := ParseHashkey(key)
				assert.NoError(t, err)
				assert.Equal(t, tc.globalPrefix, gp)
				assert.Equal(t, tc.appID, a)
				assert.Equal(t, tc.userID, u)
				assert.Equal(t, tc.prefix, p)
				assert.NotEmpty(t, h)
			}
		})
	}
}
