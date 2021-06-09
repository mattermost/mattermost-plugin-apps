package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanStaticPath(t *testing.T) {
	for _, tc := range []struct {
		p             string
		expectedError string
		expected      string
	}{
		{
			p:        `X/Y/Z`,
			expected: `X/Y/Z`,
		}, {
			p:        `X/Y/../Z`,
			expected: `X/Z`,
		}, {
			p:             `/X/Y/Z`,
			expectedError: `asset names may not start with a '/': invalid input`,
		}, {
			p:             `X/../../Y/Z`,
			expectedError: `bad path: "X/../../Y/Z": invalid input`,
		}, {
			p:             `X/Y/../../../Z`,
			expectedError: `bad path: "X/Y/../../../Z": invalid input`,
		}, {
			p:             `X%252F..%2F..%2525252FY`,
			expectedError: `bad path: "X/../../Y": invalid input`,
		}, {
			p:             `%2FX%2F..%2F..%2FY`,
			expectedError: `asset names may not start with a '/': invalid input`,
		}, {
			p:             `X%252f..%252f..%252fmanifest`,
			expectedError: `bad path: "X/../../manifest": invalid input`,
		},
	} {
		t.Run(tc.p, func(t *testing.T) {
			c, err := cleanStaticPath(tc.p)
			if tc.expectedError != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.expected, c)
			}
		})
	}
}
