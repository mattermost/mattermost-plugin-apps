// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestLoggable(t *testing.T) {
	var _ utils.HasLoggable = Config{}

	var simpleConfig = Config{
		PluginManifest: model.Manifest{
			Version: "v1.2.3",
		},
		BuildHashShort:      "1234567",
		BuildDate:           "date-is-just-a-string",
		MattermostCloudMode: true,
		DeveloperMode:       true,
		AllowHTTPApps:       true,
	}

	for name, test := range map[string]struct {
		In            interface{}
		ExpectedProps []interface{}
	}{
		"simple Config": {
			In: simpleConfig,
			ExpectedProps: []interface{}{
				"version", "v1.2.3",
				"commit", "1234567",
				"build_date", "date-is-just-a-string",
				"cloud_mode", "true",
				"developer_mode", "true",
				"allow_http_apps", "true",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if test.ExpectedProps != nil {
				lp, ok := test.In.(utils.HasLoggable)
				require.True(t, ok)
				require.EqualValues(t, test.ExpectedProps, lp.Loggable())
			}
		})
	}
}
