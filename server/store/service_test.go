// +build !e2e

package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashkey(t *testing.T) {
	for _, tc := range []struct {
		name                                string
		globalPrefix, namespace, prefix, id string
		expected                            string
		expectedLen                         int
	}{
		{
			name:         "long",
			globalPrefix: "MAXI",
			namespace:    "com.mattermost.testapp-with-a-rather-loooooooooong-name",
			prefix:       "the-app-chose-a-very-long-prefix",
			id:           "the-app-chose-an-even-very-longer-id-----------------",
			expected:     "MAXI$L$a+:*W7>S_k;\"08X9E$`5'</DONVG5fTTQ:B^jR&Hc+M",
			expectedLen:  50,
		},
		{
			name:         "short",
			globalPrefix: "x.",
			namespace:    "app",
			prefix:       "",
			id:           "0",
			expected:     "x.CJp-%PCWlJQ;b6rS<m^R3!m'!/@ZG`[Gt?1\"TiTJ%@i@0n",
			expectedLen:  48,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := hashkey(tc.globalPrefix, tc.namespace, tc.prefix, tc.id)
			require.Equal(t, tc.expected, r)
			require.Equal(t, tc.expectedLen, len(r))
		})
	}
}
