// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyUserJoinsChannel creates a test channel in a new test team. Bot, user
// and user2 are added as members of the team, and of the channel. User2 is then
// removed from the channel to trigger.
func notifyUserLeftChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper) apps.ExpandedContext {
			data := apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: th.ServerTestHelper.BasicUser2,
			}
			th.addTeamMember(data.Team, th.LastInstalledBotUser)
			th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			data.Channel = th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id)
			th.addChannelMember(data.Channel, th.LastInstalledBotUser)
			th.addChannelMember(data.Channel, th.ServerTestHelper.BasicUser)
			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
			return data
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject:   apps.SubjectUserLeftChannel,
				ChannelID: data.Channel.Id,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeUserFromChannel(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			ec := apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
			switch appclient.name {
			case "admin", "user", "bot":
				// Channel is fully expanded (user is a member of the channel,
				// and admin is admin).
				ec.Channel = data.Channel

			default: // user2
				// ChannelID gets expanded at the ID level even though the
				// acting user have no access to it.
				if level == apps.ExpandID {
					ec.Channel = &model.Channel{Id: data.Channel.Id, TeamId: data.Team.Id}
				}
			}
			return ec
		},
	}
}
