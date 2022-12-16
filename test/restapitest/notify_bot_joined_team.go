// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyBotJoinsTeam creates a new test team. User is added as a member of the
// team. Bot is then added to the channel to trigger.
func notifyBotJoinedTeam(_ *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper, _ *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			th.addTeamMember(team, th.ServerTestHelper.BasicUser)
			return apps.ExpandedContext{
				Team: team,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectBotJoinedTeam,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.TeamMember = th.addTeamMember(data.Team, th.LastInstalledBotUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User: th.LastInstalledBotUser,
			}

			switch appclient.name {
			case "admin", "bot", "user":
				ec.Team = data.Team
				ec.TeamMember = data.TeamMember

			default: // user2
				// TeamID gets expanded at the ID level even though user2 has no access to it.
				if level == apps.ExpandID {
					ec.Team = &model.Team{Id: data.Team.Id}
				}
			}
			return "<>/<>", ec
		},
	}
}
