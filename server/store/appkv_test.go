// +build !e2e,!app

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
			expected:  ("kv.D0ABymH0H8Mr9u3m10a-ow/DXm7v7FXNWbB31V8ph-F_Q"),
		},
		{
			namespace: "test_ns",
			prefix:    "test_prefix",
			id:        "test_id2",
			expected:  ("kv.pDHhAKgqD9f_UeD7gpEABw/H1FeDBrxovdZ95Lzh3BB8g"),
		},
	} {
		name := strings.Join([]string{tc.namespace, tc.prefix, tc.id}, "-")
		s := appKVStore{}
		t.Run(name, func(t *testing.T) {
			key := s.kvKey(tc.namespace, tc.prefix, tc.id)
			require.Equal(t, tc.expected, key)
		})
	}
}
