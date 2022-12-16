// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyBotJoinsChannel creates a test channel in a new test team. Bot, User
// and user2 are added as members of the team, Bot, User are added as a member
// of the channel. Bot is then removed from the channel to trigger.
func notifyBotLeftChannel(_ *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper, _ *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, th.LastInstalledBotUser)
			th.addTeamMember(team, th.ServerTestHelper.BasicUser)
			th.addTeamMember(team, th.ServerTestHelper.BasicUser2)

			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, team.Id)
			th.addChannelMember(channel, th.ServerTestHelper.BasicUser)
			th.addChannelMember(channel, th.LastInstalledBotUser)

			return apps.ExpandedContext{
				Team:       team,
				TeamMember: tm,
				Channel:    channel,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectBotLeftChannel,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeUserFromChannel(data.Channel, th.LastInstalledBotUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:       th.LastInstalledBotUser,
				Team:       data.Team,
				TeamMember: data.TeamMember,
			}
			switch appclient.name {
			case "admin", "user":
				// Channel is fully expanded (user is a member of the channel,
				// and admin is admin).
				ec.Channel = data.Channel

			default: // bot, user2
				// ChannelID gets expanded at the ID level even though the
				// acting user have no access to it.
				if level == apps.ExpandID {
					ec.Channel = &model.Channel{Id: data.Channel.Id, TeamId: data.Team.Id}
				}
			}
			return "<>/<>", ec
		},
	}
}
