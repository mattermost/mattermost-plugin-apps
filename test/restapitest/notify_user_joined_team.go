// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyUserJoinsTeam creates a new test team. Bot is added as a member of the
// team. User is then added to the team to trigger. Since user and user2 are not
// members of the team, they can not subscribe and are excluded from the test.
func notifyUserJoinedTeam(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		except: []appClient{
			th.asUser,
			th.asUser2,
		},
		init: func(th *Helper, _ *model.User) apps.ExpandedContext {
			data := apps.ExpandedContext{
				User: th.ServerTestHelper.BasicUser,
				Team: th.createTestTeam(),
			}
			th.addTeamMember(data.Team, th.LastInstalledBotUser)
			return data
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectUserJoinedTeam,
				TeamID:  data.Team.Id,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.TeamMember = th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User: th.ServerTestHelper.BasicUser,
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
