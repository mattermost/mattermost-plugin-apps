// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExpandLevel(t *testing.T) {
	for _, tc := range []struct {
		in                string
		optionalByDefault bool
		expected          string
		expectedError     string
	}{
		{
			in:       "",
			expected: "+",
		},
		{
			in:       "-",
			expected: "-",
		},
		{
			in:       "id",
			expected: "+id",
		},
		{
			in:       "-id",
			expected: "-id",
		},
		{
			in:       "+id",
			expected: "+id",
		},
		{
			in:                "id",
			optionalByDefault: true,
			expected:          "-id",
		},
		{
			in:                "-id",
			optionalByDefault: true,
			expected:          "-id",
		},
		{
			in:                "+id",
			optionalByDefault: true,
			expected:          "+id",
		},
		{
			in:       "summary",
			expected: "+summary",
		},
		{
			in:       "all",
			expected: "+all",
		},
		{
			in:       "none",
			expected: "+none",
		},
		{
			in:            "garbage",
			expectedError: `"garbage" is not a known expand level`,
		},
		{
			in:            "+garbage",
			expectedError: `"garbage" is not a known expand level`,
		},
		{
			in:            "-garbage",
			expectedError: `"garbage" is not a known expand level`,
		},
	} {
		t.Run(fmt.Sprintf("%s-%v", tc.in, tc.optionalByDefault), func(t *testing.T) {
			l, err := ParseExpandLevel(ExpandLevel(tc.in), tc.optionalByDefault)
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, string(l))
			}
		})
	}
}
