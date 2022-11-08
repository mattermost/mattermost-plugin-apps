// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func expandEverything(level apps.ExpandLevel) apps.Expand {
	return apps.Expand{
		App:                   level,
		ActingUser:            level,
		ActingUserAccessToken: level,
		Locale:                level,
		Channel:               level,
		ChannelMember:         level,
		Team:                  level,
		TeamMember:            level,
		Post:                  level,
		RootPost:              level,
		User:                  level,
		OAuth2App:             level,
		OAuth2User:            level,
	}
}

func forExpandClientCombinations(th *Helper, expandSet []apps.ExpandLevel, except []appClient, runf func(*Helper, apps.ExpandLevel, appClient)) {
	if len(expandSet) == 0 {
		expandSet = []apps.ExpandLevel{
			apps.ExpandNone,
			apps.ExpandID,
			apps.ExpandSummary,
			apps.ExpandAll,
		}
	}
	var appClients []appClient

	for _, appclient := range []appClient{th.asBot, th.asUser, th.asUser2, th.asAdmin} {
		include := true
		for _, exceptClient := range except {
			if exceptClient.name == appclient.name {
				include = false
				break
			}
		}
		if include {
			appClients = append(appClients, appclient)
		}
	}

	for _, level := range expandSet {
		name := string(level)
		if name == "" {
			name = "none"
		}
		th.Run("expand "+name, func(th *Helper) {
			for _, appclient := range appClients {
				th.Run(appclient.name, func(th *Helper) {
					runf(th, level, appclient)
				})
			}
		})
	}
}
