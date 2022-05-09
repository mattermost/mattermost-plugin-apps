// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseExpandLevel(t *testing.T) {
	for i, tc := range []struct {
		in            string
		def           ExpandLevel
		expected      string
		expectedError string
	}{
		// ""
		{
			in:       "",
			expected: "+",
		},
		{
			in:       "",
			def:      "+",
			expected: "+",
		},
		{
			in:       "",
			def:      "-",
			expected: "-",
		},
		{
			in:       "",
			def:      "+id",
			expected: "+id",
		},
		{
			in:       "",
			def:      "-id",
			expected: "-id",
		},

		// "-"
		{
			in:       "-",
			expected: "-",
		},
		{
			in:       "-",
			def:      "+",
			expected: "-",
		},
		{
			in:       "-",
			def:      "-",
			expected: "-",
		},
		{
			in:       "-",
			def:      "+id",
			expected: "-id",
		},
		{
			in:       "-",
			def:      "-id",
			expected: "-id",
		},

		// "+"
		{
			in:       "+",
			expected: "+",
		},
		{
			in:       "+",
			def:      "+",
			expected: "+",
		},
		{
			in:       "+",
			def:      "-",
			expected: "+",
		},
		{
			in:       "+",
			def:      "+id",
			expected: "+id",
		},
		{
			in:       "+",
			def:      "-id",
			expected: "+id",
		},

		// "all"
		{
			in:       "all",
			expected: "+all",
		},
		{
			in:       "all",
			def:      "+",
			expected: "+all",
		},
		{
			in:       "all",
			def:      "-",
			expected: "-all",
		},
		{
			in:       "all",
			def:      "+id",
			expected: "+all",
		},
		{
			in:       "all",
			def:      "-id",
			expected: "-all",
		},

		// "-all"
		{
			in:       "-all",
			expected: "-all",
		},
		{
			in:       "-all",
			def:      "+",
			expected: "-all",
		},
		{
			in:       "-all",
			def:      "-all",
			expected: "-all",
		},
		{
			in:       "-all",
			def:      "+id",
			expected: "-all",
		},
		{
			in:       "-all",
			def:      "-id",
			expected: "-all",
		},

		// "+all"
		{
			in:       "+all",
			expected: "+all",
		},
		{
			in:       "+all",
			def:      "+",
			expected: "+all",
		},
		{
			in:       "+all",
			def:      "-",
			expected: "+all",
		},
		{
			in:       "+all",
			def:      "+id",
			expected: "+all",
		},
		{
			in:       "+all",
			def:      "-id",
			expected: "+all",
		},

		{
			in:       "summary",
			expected: "+summary",
		},
		{
			in:       "id",
			expected: "+id",
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
		{
			in:            "all",
			def:           "garbage",
			expectedError: `failed to parse default expand level: "garbage" is not a known expand level`,
		},
	} {
		t.Run(fmt.Sprintf("level-%s-def-%v", tc.in, tc.def), func(t *testing.T) {
			l, err := ParseExpandLevel(tc.in, tc.def)
			if tc.expectedError != "" {
				require.Error(t, err, "%v", i)
				require.Equal(t, tc.expectedError, err.Error(), "%v", i)
			} else {
				require.NoError(t, err, "%v", i)
				require.Equal(t, tc.expected, string(l), "%v", i)
			}
		})
	}
}

func ExampleParseExpandLevel() {
	l, _ := ParseExpandLevel("all", ExpandNone.Required() /* "+none" */)
	fmt.Println(l)

	l, _ = ParseExpandLevel("", ExpandSummary.Optional() /* "-summary" */)
	fmt.Println(l)

	// Output:
	// +all
	// -summary
}

func TestParseExpandLevelInternal(t *testing.T) {
	for _, tc := range []struct {
		in            string
		expectedP     string
		expectedL     string
		expectedError string
	}{
		{},
		{
			in:        "+",
			expectedP: "+",
			expectedL: "",
		},
		{
			in:        "+all",
			expectedP: "+",
			expectedL: "all",
		},
		{
			in:        "-all",
			expectedP: "-",
			expectedL: "all",
		},
		{
			in:        "all",
			expectedP: "",
			expectedL: "all",
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			p, l, err := ExpandLevel(tc.in).parse()

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedP, string(p))
				require.Equal(t, tc.expectedL, string(l))
			}
		})
	}
}
