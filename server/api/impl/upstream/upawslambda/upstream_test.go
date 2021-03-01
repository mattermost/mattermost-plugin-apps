// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMatch(t *testing.T) {
	routes := []apps.Function{
		{
			Path:       "/",
			LambdaName: "main",
		}, {
			Path:       "/topic",
			LambdaName: "topic",
		}, {
			Path:       "/topic/subtopic/",
			LambdaName: "subtopic",
		}, {
			Path:       "/other",
			LambdaName: "other",
		},
	}

	for _, tc := range []struct {
		callPath string
		expected string
	}{
		{"/different", "main"},
		{"/topic/subtopic/and-then-some", "subtopic"},
		{"/topic/other/and-then-some", "topic"},
		{"/other/and-then-some", "other"},
	} {
		t.Run(tc.callPath, func(t *testing.T) {
			matched := match(tc.callPath, routes)
			assert.Equal(t, tc.expected, matched)
		})
	}
}
