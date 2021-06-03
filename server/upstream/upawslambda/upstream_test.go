// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMatch(t *testing.T) {
	lambdas := []apps.AWSLambdaFunction{
		{
			Path: "/topic",
			Name: "topic",
		}, {
			Path: "/topic/subtopic/",
			Name: "subtopic",
		}, {
			Path: "/other",
			Name: "other",
		}, {
			Path: "/",
			Name: "main",
		},
	}

	for _, tc := range []struct {
		callPath string
		expected string
	}{
		{"/different", "testID_v00-00-000_main"},
		{"/topic/subtopic/and-then-some", "testID_v00-00-000_subtopic"},
		{"/topic/other/and-then-some", "testID_v00-00-000_topic"},
		{"/other/and-then-some", "testID_v00-00-000_other"},
	} {
		t.Run(tc.callPath, func(t *testing.T) {
			matched := match(tc.callPath, &apps.Manifest{
				AppID:     "testID",
				Version:   "v00.00.000",
				AWSLambda: lambdas,
			})
			assert.Equal(t, tc.expected, matched)
		})
	}
}
