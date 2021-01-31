// +build !e2e

package appservices

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/stretchr/testify/require"
)

// nolint:gosec

func TestKVKey(t *testing.T) {
	for _, tc := range []struct {
		namespace, prefix, id string
		expected              string
	}{
		{
			expected: "",
		},
		{
			namespace: "test_ns",
			id:        "test_id",
			expected:  ("0f4001ca61f41fc32bf6ede6d746bea3//2e06cda4c3c0d4a2a42058f74641546a")[:model.KEY_VALUE_KEY_MAX_RUNES],
		},
		{
			namespace: "test_ns",
			prefix:    "test_prefix",
			id:        "test_id",
			expected:  ("0f4001ca61f41fc32bf6ede6d746bea3/test_prefix/2e06cda4c3c0d4a2a42058f74641546a")[:model.KEY_VALUE_KEY_MAX_RUNES],
		},
	} {
		name := strings.Join([]string{tc.namespace, tc.prefix, tc.id}, "_")
		t.Run(name, func(t *testing.T) {
			key := kvKey(tc.namespace, tc.prefix, tc.id)
			require.Equal(t, tc.expected, key)
		})
	}
}
