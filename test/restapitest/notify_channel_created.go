// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyChannelCreated creates a test channel in a new test team. User and
// user2 are added as members of the team, so they and the admin can subscribe
// and receive notifications. Bot is not added to the team, so it can not
// subscribe.
func notifyChannelCreated(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		appClients: []appClient{
			th.asUser,
			th.asUser2,
			th.asAdmin,
		},
		init: func(th *Helper) apps.ExpandedContext {
			// create test team, and make "user" a member (but not bot, nor user2)
			team := th.createTestTeam()
			th.addTeamMember(team, th.ServerTestHelper.BasicUser)
			th.addTeamMember(team, th.ServerTestHelper.BasicUser2)
			return apps.ExpandedContext{
				Team: team,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectChannelCreated,
				TeamID:  data.Team.Id,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				Channel: th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id),
			}
		},
		expected: func(th *Helper, level apps.ExpandLevel, cl appClient, data apps.ExpandedContext) apps.ExpandedContext {
			// only user, user2 and admin can get here, bit wouldn't be able to
			// subscribe since it was not added to the team in init.
			switch cl.name {
			case "admin":
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id), // data.Channel,
					ChannelMember: th.getChannelMember(data.Channel.Id, cl.expectedActingUser.Id),
					Team:          th.getTeam(data.Channel.TeamId),
					TeamMember:    th.getTeamMember(data.Channel.TeamId, cl.expectedActingUser.Id),
				}

			default: // user, user2
				ec := apps.ExpandedContext{
					Team: th.getTeam(data.Channel.TeamId),
					// ChannelMember: th.getChannelMember(data.Channel.Id, cl.expectedActingUser.Id),
					TeamMember: th.getTeamMember(data.Channel.TeamId, cl.expectedActingUser.Id),
				}
				if level == apps.ExpandID {
					ec.Channel = th.getChannel(data.Channel.Id) // data.Channel,
				}
				return ec
			}
		},
	}
}
