// +build !e2e

package store

import (
	"strings"
	"testing"

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
			id:        "test_id1",
			expected:  ("__E17qMxGaOyuJqYQs5PB8PtHS7A/tMjpGjp2h7PlMhUBjLzjO"),
		},
		{
			namespace: "test_ns",
			prefix:    "test_prefix",
			id:        "test_id2",
			expected:  ("_iGXV5w_xPrIPpq84ntMkm_99yls/A9CtwkTHxKvq_DBgdNy6C"),
		},
	} {
		name := strings.Join([]string{tc.namespace, tc.prefix, tc.id}, "_")
		t.Run(name, func(t *testing.T) {
			key := kvKey(tc.namespace, tc.prefix, tc.id)
			require.Equal(t, tc.expected, key)
		})
	}
}
