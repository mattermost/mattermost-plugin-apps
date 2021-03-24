// +build !e2e

package appservices

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
			id:        "test_id",
			expected:  ("_E17qMxGaOyuJqYQs5PB8PtHS7A/VBevAGLPmHSVthG1nH7DdU"),
		},
		{
			namespace: "test_ns",
			prefix:    "test_prefix",
			id:        "test_id",
			expected:  ("iGXV5w_xPrIPpq84ntMkm_99yls/VBevAGLPmHSVthG1nH7DdU"),
		},
	} {
		name := strings.Join([]string{tc.namespace, tc.prefix, tc.id}, "_")
		t.Run(name, func(t *testing.T) {
			key := kvKey(tc.namespace, tc.prefix, tc.id)
			require.Equal(t, tc.expected, key)
		})
	}
}
