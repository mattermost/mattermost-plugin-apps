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
func notifyAnyUserJoinedTheTeam(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectUserJoinedTeam,
				TeamID:  data.Team.Id,
			}
		},
		except: []appClient{
			th.asUser2,
		},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			joiningUser := th.ServerTestHelper.BasicUser2
			team := th.createTestTeam()
			th.addTeamMember(team, user)
			return apps.ExpandedContext{
				User: joiningUser,
				Team: team,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
		},
	}
}

func notifySubscriberJoinedAnyTeam(th *Helper) *notifyTestCase {
	return notifyTheUserJoinedAnyTeam(th, apps.SubjectUserJoinedTeam, []appClient{th.asAdmin})
}

func notifyBotJoinedAnyTeam(th *Helper) *notifyTestCase {
	return notifyTheUserJoinedAnyTeam(th, apps.SubjectBotJoinedTeamDeprecated, []appClient{th.asAdmin, th.asUser2, th.asUser})
}

func notifyTheUserJoinedAnyTeam(th *Helper, subject apps.Subject, except []appClient) *notifyTestCase {
	return &notifyTestCase{
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: subject,
			}
		},
		except: except,
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			return apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: user,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
		},
	}
}
