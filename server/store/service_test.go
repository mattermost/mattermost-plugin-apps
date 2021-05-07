// +build !e2e

package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashkey(t *testing.T) {
	s := NewService(nil, nil)
	for _, tc := range []struct {
		name                                string
		globalPrefix, botUserID, prefix, id string
		expectedError                       string
		expected                            string
		expectedLen                         int
	}{
		{
			name:         "long",
			globalPrefix: ".X",
			botUserID:    "botUserIDis26bytes90123456",
			prefix:       "",
			id:           "the-app-chose-an-even-very-longer-id-----------------",
			expected:     ".XbotUserIDis26bytes90123456  DONVG5fTTQ:B^jR&Hc+M",
			expectedLen:  50,
		},
		{
			name:         "one byte prefix",
			globalPrefix: ".X",
			botUserID:    "botUserIDis26bytes90123456",
			prefix:       "A",
			id:           "id",
			expected:     ".XbotUserIDis26bytes90123456A nEDMRpHe-;lXX_IkLt)+",
			expectedLen:  50,
		},
		{
			name:         "two byte prefix",
			globalPrefix: ".X",
			botUserID:    "botUserIDis26bytes90123456",
			prefix:       "AB",
			id:           "id",
			expected:     ".XbotUserIDis26bytes90123456ABnEDMRpHe-;lXX_IkLt)+",
			expectedLen:  50,
		},
		{
			name:         "short",
			globalPrefix: ".X",
			botUserID:    "botUserIDis26bytes90123456",
			prefix:       "",
			id:           "0",
			expected:     ".XbotUserIDis26bytes90123456  @ZG`[Gt?1\"TiTJ%@i@0n",
			expectedLen:  50,
		},
		{
			name:          "error empty globalPrefix",
			globalPrefix:  "",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix does not start with a dot",
			globalPrefix:  "AB",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "AB" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix too short",
			globalPrefix:  ".",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix "." is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error globalPrefix too long",
			globalPrefix:  ".TOOLONG",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "A",
			id:            "id",
			expectedError: `global prefix ".TOOLONG" is not 2 ASCII characters starting with a '.'`,
		},
		{
			name:          "error botUserID too short",
			globalPrefix:  ".X",
			botUserID:     "TOOSHORT",
			prefix:        "A",
			id:            "id",
			expectedError: `botUserID "TOOSHORT" must be exactly 26 ASCII characters`,
		},
		{
			name:          "error botUserID too long",
			globalPrefix:  ".X",
			botUserID:     "botUserIDis26bytes90123456TOOLONG",
			prefix:        "A",
			id:            "id",
			expectedError: `botUserID "botUserIDis26bytes90123456TOOLONG" must be exactly 26 ASCII characters`,
		},
		{
			name:          "error id empty",
			globalPrefix:  ".X",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "A",
			id:            "",
			expectedError: `key must not be empty`,
		},
		{
			name:          "error prefix too long",
			globalPrefix:  ".X",
			botUserID:     "botUserIDis26bytes90123456",
			prefix:        "ABC",
			id:            "id",
			expectedError: `prefix "ABC" is longer than the limit of 2 ASCII characters`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			key, err := s.hashkey(tc.globalPrefix, tc.botUserID, tc.prefix, tc.id)
			if tc.expectedError != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, key)
				require.Equal(t, tc.expectedLen, len(key))

				gp, b, p, h, _ := parseHashkey(key)
				require.Equal(t, tc.globalPrefix, gp)
				require.Equal(t, tc.botUserID, b)
				require.Equal(t, tc.prefix, p)
				require.NotEmpty(t, h)
			}
		})
	}
}
