// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-server/v6/model"
)

// notifyChannelCreated creates a test channel in a new test team. User and
// user2 are added as members of the team, so they and the admin can subscribe
// and receive notifications. Bot is not added to the team, so it can not
// subscribe.
func notifyChannelCreated(th *Helper) *notifyTestCase {
	return &notifyTestCase{
		except: []appClient{
			th.asBot,
		},
		init: func(th *Helper, _ *model.User) apps.ExpandedContext {
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
			data.Channel = th.createTestChannel(th.ServerTestHelper.SystemAdminClient, data.Team.Id)
			return data
		},
		expected: func(th *Helper, level apps.ExpandLevel, appclient appClient, data apps.ExpandedContext) (apps.Subject, apps.ExpandedContext) {
			// only user, user2 and admin can get here, bit wouldn't be able to
			// subscribe since it was not added to the team in init.
			switch appclient.name {
			case "admin":
				return "<>/<>", apps.ExpandedContext{
					Channel:       data.Channel,
					ChannelMember: th.getChannelMember(data.Channel.Id, appclient.expectedActingUser.Id),
					Team:          data.Team,
					TeamMember:    th.getTeamMember(data.Channel.TeamId, appclient.expectedActingUser.Id),
				}

			default: // user, user2, bot
				ec := apps.ExpandedContext{
					Team:       data.Team,
					TeamMember: th.getTeamMember(data.Channel.TeamId, appclient.expectedActingUser.Id),
				}
				if level == apps.ExpandID {
					ec.Channel = data.Channel
				}
				return "<>/<>", ec
			}
		},
	}
}
