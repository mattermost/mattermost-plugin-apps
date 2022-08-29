package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCleanStaticURL(t *testing.T) {
	for _, tc := range []struct {
		p             string
		expectedError string
		expected      string
	}{
		{
			p:        `X/Y/Z`,
			expected: `X/Y/Z`,
		}, {
			p:        `https://test.t/X/Y/Z`,
			expected: `https://test.t/X/Y/Z`,
		}, {
			p:        `X/Y/../Z`,
			expected: `X/Z`,
		}, {
			p:        `http://test.t:8080/X/Y/../Z`,
			expected: `http://test.t:8080/X/Z`,
		}, {
			p:        `/X/Y/Z`,
			expected: `./X/Y/Z`,
		}, {
			p:        `////X///Y////Z///`,
			expected: `./X/Y/Z`,
		}, {
			p:        `https://test.t//X/Y/Z`,
			expected: `https://test.t/X/Y/Z`,
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
			p:        `%2FX%2F..%2F..%2FY`,
			expected: `./Y`,
		}, {
			p:        `/X/../../Y`,
			expected: `./Y`,
		}, {
			p:        `https://test.t/X/../../Y`,
			expected: `https://test.t/Y`,
		}, {
			p:             `X%252f..%252f..%252fmanifest`,
			expectedError: `bad path: "X/../../manifest": invalid input`,
		},
	} {
		t.Run(tc.p, func(t *testing.T) {
			c, err := CleanStaticURL(tc.p)
			if tc.expectedError != "" {
				require.NotNil(t, err, "expected: %s", tc.expectedError)
				require.Equal(t, tc.expectedError, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.expected, c)
			}
		})
	}
}
