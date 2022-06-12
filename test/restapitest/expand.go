// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

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

func forExpandClientCombinations(th *Helper, appBotUser *model.User, expandSet []apps.ExpandLevel, appClients []appClient, runf func(*Helper, apps.ExpandLevel, appClient)) {
	if len(expandSet) == 0 {
		expandSet = []apps.ExpandLevel{
			apps.ExpandNone,
			apps.ExpandID,
			apps.ExpandSummary,
			apps.ExpandAll,
		}
	}
	if len(appClients) == 0 {
		appClients = []appClient{th.asBot, th.asUser, th.asUser2, th.asAdmin}
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
