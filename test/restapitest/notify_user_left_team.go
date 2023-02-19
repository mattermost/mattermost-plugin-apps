// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyUserLeftTeam creates a new test team. User, user2 and bot are added as members of the
// team. User2 is then removed from the team to trigger.
func notifyAnyUserLeftTheTeam(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		except: []appClient{th.asUser2},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			joiningUser := th.ServerTestHelper.BasicUser2
			tm := th.addTeamMember(team, joiningUser)
			th.addTeamMember(team, user)
			return apps.ExpandedContext{
				Team:       team,
				User:       joiningUser,
				TeamMember: tm,
			}
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
			return apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
		},
	}
}

func notifyBotLeftAnyTeam(th *Helper) *notifyTestCase {
	return notifyTheUserLeftAnyTeam(th, apps.SubjectBotLeftTeam, []appClient{th.asAdmin, th.asUser, th.asUser2})
}

func notifySubscriberLeftAnyTeam(th *Helper) *notifyTestCase {
	return notifyTheUserLeftAnyTeam(th, apps.SubjectUserLeftTeam, []appClient{th.asAdmin})
}

func notifyTheUserLeftAnyTeam(th *Helper, subject apps.Subject, except []appClient) *notifyTestCase {
	return &notifyTestCase{
		except: except,
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, user)
			return apps.ExpandedContext{
				Team:       team,
				TeamMember: tm,
				User:       user,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{Subject: subject}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeTeamMember(data.Team, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			ec := apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
			ec.TeamMember.Roles = ""
			return ec
		},
	}
}
