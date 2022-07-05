// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyUserLeftTeam creates a new test team. User, user2 and bot are added as members of the
// team. User2 is then removed from the team to trigger.
func notifyUserLeftTeam(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper) apps.ExpandedContext {
			data := apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: th.ServerTestHelper.BasicUser2,
			}
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)
			th.addTeamMember(data.Team, th.LastInstalledBotUser)
			return data
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectUserLeftTeam,
				TeamID:  data.Team.Id,
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
			case "admin", "user", "bot":
				ec.Team = data.Team
				ec.TeamMember = th.getTeamMember(data.Team.Id, data.User.Id)
			default: // user2
				// TeamID gets expanded at the ID level even though user2 has no access to it.
				if level == apps.ExpandID {
					ec.Team = &model.Team{Id: data.Team.Id}
				}
			}
			return ec
		},
	}
}
