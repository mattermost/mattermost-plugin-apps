// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyBotLeftTeam creates a new test team. User and Bot are added as members of the
// team. Bot is then removed from the team to trigger.
func notifyBotLeftTeam(_ *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper) apps.ExpandedContext {
			th.Skip("https://mattermost.atlassian.net/browse/MM-48497")
			data := apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: th.LastInstalledBotUser,
			}
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)
			return data
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectBotLeftTeam,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeTeamMember(data.Team, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			ec := apps.ExpandedContext{
				User: data.User,
			}

			switch appclient.name {
			case "admin", "user":
				ec.Team = data.Team
				ec.TeamMember = th.getTeamMember(data.Team.Id, th.LastInstalledApp.BotUserID)

			default: // user2, bot
				// TeamID gets expanded at the ID level even though user2 has no access to it.
				if level == apps.ExpandID {
					ec.Team = &model.Team{Id: data.Team.Id}
				}
			}
			return ec
		},
	}
}
