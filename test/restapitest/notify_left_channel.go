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
func notifyAnyUserLeftTheChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject:   apps.SubjectUserLeftChannel,
				ChannelID: data.Channel.Id,
			}
		},
		except: []appClient{th.asUser2},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			data := apps.ExpandedContext{
				User: th.ServerTestHelper.BasicUser2,
			}
			data.Team = th.createTestTeam()
			data.Channel = th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id)
			data.TeamMember = th.addTeamMember(data.Team, data.User)
			th.addTeamMember(data.Team, user)
			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
			th.addChannelMember(data.Channel, user)
			return data
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeUserFromChannel(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
				Channel:    data.Channel,
			}
			return apps.SubjectUserLeftChannel, ec
		},
	}
}

func notifySubscriberLeftAnyChannel(th *Helper) *notifyTestCase {
	return notifyTheUserLeftAnyChannel(th, apps.SubjectUserLeftChannel, []appClient{th.asAdmin})
}

func notifyBotLeftAnyChannel(th *Helper) *notifyTestCase {
	return notifyTheUserLeftAnyChannel(th, apps.SubjectBotLeftChannel_Deprecated, []appClient{th.asUser, th.asUser2, th.asAdmin})
}

func notifyTheUserLeftAnyChannel(th *Helper, subject apps.Subject, except []appClient) *notifyTestCase {
	return &notifyTestCase{
		except: except,
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: subject,
			}
		},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, user)
			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, team.Id)
			cm := th.addChannelMember(channel, user)

			return apps.ExpandedContext{
				Team:          team,
				TeamMember:    tm,
				Channel:       channel,
				ChannelMember: cm,
				ActingUser:    user,
				User:          user,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			th.removeUserFromChannel(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:          data.User,
				ActingUser:    data.ActingUser,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
			}
			return subject, ec
		},
	}
}
