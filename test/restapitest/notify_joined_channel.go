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
	return notifyBotJoinedChannelImpl(th, apps.SubjectBotJoinedChannel, apps.SubjectSelfJoinedChannel)
}

func notifyBotJoinedChannelLegacy(th *Helper) *notifyTestCase {
	return notifyBotJoinedChannelImpl(th, apps.SubjectBotJoinedChannel, apps.SubjectBotJoinedChannel)
}

func notifyBotJoinedChannelImpl(th *Helper, subjectIn, subjectOut apps.Subject, testFlag bool, except ) *notifyTestCase {
	return &notifyTestCase{
		except: []appClient{th.asAdmin, th.asUser, th.asUser2},
		init: func(th *Helper, user *model.User) apps.ExpandedContext {
			team := th.createTestTeam()
			tm := th.addTeamMember(team, th.LastInstalledBotUser)
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
				ActingUser:    th.LastInstalledBotUser,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				User:          th.LastInstalledBotUser,
			}

			// Remap the Subject, will receive SubjectSelfJoinedChannel, SubjectBotJoinedChannel is deprecvated.
			return subjectOut, ec
		},
	}
}

// notifyBotJoinsChannel creates a test channel in a new test team. Bot, User
// and user2 are added as members of the team, and User is added as a member of
// the channel. Bot is then added to the channel to trigger.
func notifyBotJoinedChannelLegacy(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		useTestSubscribe: true,
		except:           []appClient{th.asAdmin, th.asUser, th.asUser2},
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
				Subject: apps.SubjectBotJoinedChannel,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.ActingUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:          th.LastInstalledBotUser,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
			}
			return apps.SubjectBotJoinedChannel, ec
		},
	}
}

func notifySelfJoinedChannel(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		// admin creates the channel, so can not be added to it again, not
		// easily. It should be covered by user tests though.
		except: []appClient{th.asAdmin},
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
				Subject: apps.SubjectSelfJoinedChannel,
			}
		},
		trigger: func(th *Helper, data apps.ExpandedContext) apps.ExpandedContext {
			data.ChannelMember = th.addChannelMember(data.Channel, data.ActingUser)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			ec := apps.ExpandedContext{
				User:          data.ActingUser,
				ActingUser:    data.ActingUser,
				Team:          data.Team,
				TeamMember:    data.TeamMember,
				Channel:       data.Channel,
				ChannelMember: data.ChannelMember,
			}

			return apps.SubjectSelfJoinedChannel, ec
		},
	}
}
