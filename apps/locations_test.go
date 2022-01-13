// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestInLocation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		out, in  apps.Location
		expected bool
	}{
		{
			name:     "Happy1",
			out:      "/a",
			in:       "/a/b",
			expected: true,
		},
		{
			name:     "Happy2",
			out:      "/a",
			in:       "/a-b",
			expected: true,
		},
		{
			name: "Not1",
			out:  "/a",
			in:   "/b",
		},
		{
			name: "Not2",
			out:  "/a/b",
			in:   "/a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.in.In(tc.out))
		})
	}
}
