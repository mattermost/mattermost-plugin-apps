// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyAnyUserJoinedTheChannel tests SubjectUserJoinedChannel with a specific
// ChannelID. It creates a test channel in a new test team. The acting test user
// is added as a channel member. BasicUser2 is then added to the channel to
// trigger.
func notifyAnyUserJoinedTheChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject:   apps.SubjectUserJoinedChannel,
				ChannelID: data.Channel.Id,
			}
		},
		except: []appClient{th.asUser2},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			data := apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: th.ServerTestHelper.BasicUser2,
			}
			data.TeamMember = th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser2)
			data.Channel = th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id)
			th.addTeamMember(data.Team, user)
			th.addChannelMember(data.Channel, user)
			return data
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				User:          data.User,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
			}
		},
	}
}

func notifySubscriberJoinedAnyChannel(th *Helper) *notifyTestCase {
	return notifyTheUserJoinedAnyChannel(th, apps.SubjectUserJoinedChannel, []appClient{th.asAdmin})
}

func notifyBotJoinedAnyChannel(th *Helper) *notifyTestCase {
	return notifyTheUserJoinedAnyChannel(th, apps.SubjectBotJoinedChannelDeprecated, []appClient{th.asUser, th.asUser2, th.asAdmin})
}

// notifyTheUserJoinedAnyChannelImpl tests SubjectUserJoinedChannel with no
// ChannelID, and SubjectBotJoinedChannel. It creates a test channel in a new
// test team. Bot and user2 are added as members of the team and channel.  User
// is then added to the channel to trigger.
func notifyTheUserJoinedAnyChannel(th *Helper, subject apps.Subject, except []appClient) *notifyTestCase {
	return &notifyTestCase{
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: subject,
			}
		},
		except: except,
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, user)
			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, team.Id)

			return apps.ExpandedContext{
				Team:       team,
				TeamMember: tm,
				Channel:    channel,
				User:       user,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) apps.ExpandedContext {
			return apps.ExpandedContext{
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				User:          data.User,
			}
		},
	}
}
