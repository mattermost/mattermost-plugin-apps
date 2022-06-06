// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandLevel(t *testing.T) {
	for _, tc := range []struct {
		in               ExpandLevel
		expectedRequired bool
		expectedLevel    ExpandLevel
		expectedError    string
	}{
		// ""
		{
			in:            "",
			expectedLevel: ExpandNone,
		},
		{
			in:               "+",
			expectedLevel:    ExpandNone,
			expectedRequired: true,
		},
		{
			in:            "id",
			expectedLevel: ExpandID,
		},
		{
			in:               "+id",
			expectedLevel:    ExpandID,
			expectedRequired: true,
		},
		{
			in:            "summary",
			expectedLevel: ExpandSummary,
		},
		{
			in:               "+summary",
			expectedLevel:    ExpandSummary,
			expectedRequired: true,
		},
		{
			in:            "all",
			expectedLevel: ExpandAll,
		},
		{
			in:               "+all",
			expectedLevel:    ExpandAll,
			expectedRequired: true,
		},
		{
			in:            "garbage",
			expectedError: `"garbage" is not a known expand level`,
		},
		{
			in:            "+garbage",
			expectedError: `"garbage" is not a known expand level`,
		},
	} {
		name := string(tc.in)
		if name == "" {
			name = "-none-"
		}
		t.Run(name, func(t *testing.T) {
			required, l, err := ParseExpandLevel(tc.in)
			if tc.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedLevel, l)
				require.Equal(t, tc.expectedRequired, required)
			}
		})
	}
}
