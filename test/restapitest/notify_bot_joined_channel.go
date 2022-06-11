// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func notifyBotJoinedChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		init: func(th *Helper) apps.ExpandedContext {
			team := th.createTestTeam()
			th.addTeamMember(team, th.AppBotUser)
			th.addTeamMember(team, th.ServerTestHelper.BasicUser)
			th.addTeamMember(team, th.ServerTestHelper.BasicUser2)

			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, team.Id)
			th.addChannelMember(channel, th.ServerTestHelper.BasicUser)

			return apps.ExpandedContext{
				Team:    team,
				Channel: channel,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: apps.SubjectBotJoinedChannel,
				TeamID:  data.Team.Id,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, th.AppBotUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, cl appClient, data apps.ExpandedContext) apps.ExpandedContext {
			switch cl.name {
			case "admin", "bot":
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id),
					ChannelMember: th.getChannelMember(data.Channel.Id, th.AppBotUser.Id),
					User:          th.getUser(th.AppBotUser.Id),
				}

			case "user":
				return apps.ExpandedContext{
					Channel:       th.getChannel(data.Channel.Id),
					ChannelMember: th.getChannelMember(data.Channel.Id, th.AppBotUser.Id),
					User:          th.getUser(th.AppBotUser.Id),
				}

			default: // user2
				ec := apps.ExpandedContext{
					User: th.getUser(th.AppBotUser.Id),
				}
				// ChannelID gets expanded at the ID level even though user2 has no access to it.
				if level == apps.ExpandID {
					ec.Channel = &model.Channel{
						Id: data.Channel.Id,
					}
				}
				return ec
			}
		},
	}
}
