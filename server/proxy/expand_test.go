package proxy

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestExpand(t *testing.T) {
	app := apps.App{
		BotUserID:   "botid",
		BotUsername: "botusername",
		DeployType:  apps.DeployBuiltin,
		Manifest: apps.Manifest{
			AppID:       apps.AppID("app1"),
			DisplayName: "App 1",
		},
	}

	userID := "user4567890123456789012345"
	channelID := "channel7890123456789012345"
	teamID := "team4567890123456789012345"
	userChannelRoles := "user_channel_roles"
	channelMember := &model.ChannelMember{
		UserId:    userID,
		ChannelId: channelID,
		Roles:     userChannelRoles,
	}
	teamMember := &model.TeamMember{
		UserId: userID,
		TeamId: teamID,
		Roles:  userChannelRoles,
	}

	type TC struct {
		base              apps.Context
		expectClientCalls func(*mock_mmclient.MockClient)
		expect            map[apps.ExpandLevel]interface{} // string for err.Error, or apps.ExpandedContext for success
	}

	for _, field := range []struct {
		name string
		tcs  map[string]TC
	}{
		{
			name: "channel_member",
			tcs: map[string]TC{
				"happy": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetChannelMember(channelID, userID).Times(2).Return(channelMember, nil)
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{
							AppPath:       "/apps/app1",
							BotUserID:     "botid",
							ActingUser:    &model.User{Id: userID},
							ChannelMember: channelMember,
						},
						apps.ExpandSummary: apps.ExpandedContext{
							AppPath:       "/apps/app1",
							BotUserID:     "botid",
							ActingUser:    &model.User{Id: userID},
							ChannelMember: channelMember,
						},
						apps.ExpandDefault:            apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
						apps.ExpandNone:               apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
						apps.ExpandLevel("jibberish"): "failed to expand channel membership: unknown expand type \"jibberish\"",
					},
				},
				"no user ID": {
					base: apps.Context{
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid"},
					},
				},
				"no channel ID": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
					},
				},
				"API error": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetChannelMember(channelID, userID).Times(1).Return(nil, errors.New("ERROR"))
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: "failed to expand channel membership: failed to get channel membership: ERROR",
					},
				},
			},
		},
		{
			name: "team_member",
			tcs: map[string]TC{
				"happy": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetTeamMember(teamID, userID).Times(2).Return(teamMember, nil)
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{
							AppPath:    "/apps/app1",
							BotUserID:  "botid",
							ActingUser: &model.User{Id: userID},
							TeamMember: teamMember,
						},
						apps.ExpandSummary: apps.ExpandedContext{AppPath: "/apps/app1",
							BotUserID:  "botid",
							ActingUser: &model.User{Id: userID},
							TeamMember: teamMember,
						},
						apps.ExpandDefault:            apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
						apps.ExpandNone:               apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
						apps.ExpandLevel("jibberish"): "failed to expand team membership: unknown expand type \"jibberish\"",
					},
				},
				"no user ID": {
					base: apps.Context{
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid"},
					},
				},
				"no team ID": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: apps.ExpandedContext{AppPath: "/apps/app1", BotUserID: "botid", ActingUser: &model.User{Id: userID}},
					},
				},
				"API error": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetTeamMember(teamID, userID).Times(1).Return(nil, errors.New("ERROR"))
					},
					expect: map[apps.ExpandLevel]interface{}{
						apps.ExpandAll: "failed to expand team membership: failed to get team membership team4567890123456789012345: ERROR",
					},
				},
			},
		},
	} {
		t.Run(field.name, func(t *testing.T) {
			for name, tc := range field.tcs {
				t.Run(name, func(t *testing.T) {
					conf := config.NewTestConfigService(nil).WithMattermostConfig(model.Config{
						ServiceSettings: model.ServiceSettings{
							SiteURL: model.NewString("test.mattermost.com"),
						},
					})
					ctrl := gomock.NewController(t)
					client := mock_mmclient.NewMockClient(ctrl)
					if tc.expectClientCalls != nil {
						tc.expectClientCalls(client)
					}
					p := &Proxy{
						conf:                 conf,
						expandClientOverride: client,
					}

					for level, expected := range tc.expect {
						t.Run(string(level), func(t *testing.T) {
							clone := tc.base
							expandData := fmt.Sprintf(`{"%s":"%s"}`, field.name, level)
							e := apps.Expand{}
							err := json.Unmarshal([]byte(expandData), &e)
							require.NoError(t, err)

							r := incoming.NewRequest(conf.MattermostAPI(), conf, utils.NewTestLogger(), nil)
							cc, err := p.expandContext(r, app, &clone, &e)
							if err != nil {
								require.EqualValues(t, expected, err.Error())
							} else {
								require.EqualValues(t, expected, cc.ExpandedContext)
							}
						})
					}
				})
			}
		})
	}
}
