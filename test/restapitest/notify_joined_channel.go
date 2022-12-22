// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// notifyUserJoinsChannel creates a test channel in a new test team. Bot and
// user are added as members of the team; Bot is added as a member of the
// channel. User is then added to the channel to trigger. Since user and user2
// are not members of the channel, thay can not subscribe and is excluded from
// the test.
func notifyUserJoinedChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		except: []appClient{
			th.asUser,
			th.asUser2,
		},
		init: func(th *Helper, _ *model.User) apps.ExpandedContext {
			data := apps.ExpandedContext{
				Team: th.createTestTeam(),
				User: th.ServerTestHelper.BasicUser,
			}
			th.addTeamMember(data.Team, th.LastInstalledBotUser)
			th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)
			data.TeamMember = th.addTeamMember(data.Team, th.ServerTestHelper.BasicUser)

			data.Channel = th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id)
			th.addChannelMember(data.Channel, th.LastInstalledBotUser)
			return data
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject:   apps.SubjectUserJoinedChannel,
				ChannelID: data.Channel.Id,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.User)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:       data.User,
				Team:       data.Team,
				TeamMember: data.TeamMember,
				Channel:    &model.Channel{Id: data.Channel.Id, TeamId: data.Team.Id},
			}
			switch appclient.name {
			case "admin", "bot", "user":
				ec.Channel = data.Channel
				ec.ChannelMember = data.ChannelMember
			}
			return apps.SubjectUserJoinedChannel, ec
		},
	}
}

func notifyBotJoinedChannelRemapped(th *Helper) *notifyTestCase {
	return notifyBotJoinedChannelImpl(th, apps.SubjectBotJoinedChannel, apps.SubjectSelfJoinedChannel, false, []appClient{th.asAdmin, th.asUser, th.asUser2})
}

func notifyBotJoinedChannelLegacy(th *Helper) *notifyTestCase {
	return notifyBotJoinedChannelImpl(th, apps.SubjectBotJoinedChannel, apps.SubjectBotJoinedChannel, true, []appClient{th.asAdmin, th.asUser, th.asUser2})
}

func notifySelfJoinedChannel(th *Helper) *notifyTestCase {
	return notifyBotJoinedChannelImpl(th, apps.SubjectSelfJoinedChannel, apps.SubjectSelfJoinedChannel, false, []appClient{th.asAdmin})
}

func notifyBotJoinedChannelImpl(th *Helper, subjectIn, subjectOut apps.Subject, testFlag bool, except []appClient) *notifyTestCase {
	return &notifyTestCase{
		useTestSubscribe: testFlag,
		except:           except,
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, user)
			channel := th.createTestChannel(th.ServerTestHelper.SystemAdminClient, team.Id)

			return apps.ExpandedContext{
				Team:       team,
				TeamMember: tm,
				Channel:    channel,
				ActingUser: user,
			}
		},
		event: func(th *Helper, data apps.ExpandedContext) apps.Event {
			return apps.Event{
				Subject: subjectIn,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.ActingUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				ActingUser:    data.ActingUser,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				User:          data.ActingUser,
			}

			return subjectOut, ec
		},
	}
}
