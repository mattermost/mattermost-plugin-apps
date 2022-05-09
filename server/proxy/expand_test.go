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

	channelMemberIDOnly := model.ChannelMember{
		UserId:    userID,
		ChannelId: channelID,
	}

	channelMember := channelMemberIDOnly
	channelMember.Roles = userChannelRoles

	teamMemberIDOnly := model.TeamMember{
		UserId: userID,
		TeamId: teamID,
	}

	teamMember := teamMemberIDOnly
	teamMember.Roles = userChannelRoles

	actingUserIDOnly := &model.User{
		Id: userID,
	}

	actingUserSummary := &model.User{
		BotDescription: "test bot",
		DeleteAt:       2000,
		Email:          "test@test.test",
		FirstName:      "test first name",
		Id:             userID,
		IsBot:          true,
		LastName:       "test last name",
		Locale:         "test locale",
		Nickname:       "test nickname",
		Roles:          "test roles",
		Username:       "test_username",
	}

	// fields that are only expanded for "all"
	actingUser := func() *model.User {
		u := *actingUserSummary
		u.UpdateAt = 1000
		u.CreateAt = 1000
		u.LastActivityAt = 1500
		u.Props = model.StringMap{
			"test_prop": "test value",
		}
		return &u
	}

	type TC struct {
		base              apps.Context
		expectClientCalls func(*mock_mmclient.MockClient)
		expect            map[string]interface{} // string for err.Error, or apps.ExpandedContext for success
	}

	expected := func(ec apps.ExpandedContext) apps.ExpandedContext {
		ec.MattermostSiteURL = "https://test.mattermost.test"
		ec.DeveloperMode = true
		ec.AppPath = "/apps/app1"
		ec.BotUserID = "botid"
		ec.App = &apps.App{
			BotUserID: "botid",
		}
		return ec
	}

	for _, field := range []struct {
		name string
		tcs  map[string]TC
	}{
		{
			name: "acting_user",
			tcs: map[string]TC{
				"happy with API GetUser": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetUser(userID).Times(1).Return(actingUser(), nil)
					},
					expect: map[string]interface{}{
						"-all":     expected(apps.ExpandedContext{ActingUser: actingUser()}),
						"-summary": expected(apps.ExpandedContext{ActingUser: actingUserSummary}),
						"+all":     expected(apps.ExpandedContext{ActingUser: actingUser()}),
						"+summary": expected(apps.ExpandedContext{ActingUser: actingUserSummary}),
						"all":      expected(apps.ExpandedContext{ActingUser: actingUser()}),
						"summary":  expected(apps.ExpandedContext{ActingUser: actingUserSummary}),
					},
				},
				"happy no API": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[string]interface{}{
						"-id":  expected(apps.ExpandedContext{ActingUser: actingUserIDOnly}),
						"+id":  expected(apps.ExpandedContext{ActingUser: actingUserIDOnly}),
						"id":   expected(apps.ExpandedContext{ActingUser: actingUserIDOnly}),
						"":     expected(apps.ExpandedContext{}),
						"none": expected(apps.ExpandedContext{}),
					},
				},
				"error GetUser fail": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetUser(userID).Times(1).Return(nil, utils.ErrForbidden)
					},
					expect: map[string]interface{}{
						"-all":     expected(apps.ExpandedContext{}),
						"-summary": expected(apps.ExpandedContext{}),
						"+all":     "failed to expand required acting_user: id: user4567890123456789012345: forbidden",
						"+summary": "failed to expand required acting_user: id: user4567890123456789012345: forbidden",
						"all":      "failed to expand required acting_user: id: user4567890123456789012345: forbidden",
						"summary":  "failed to expand required acting_user: id: user4567890123456789012345: forbidden",
					},
				},
				"error invalid": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[string]interface{}{
						"garbage":  `"garbage" is not a known expand level`,
						"+garbage": `"garbage" is not a known expand level`,
						"-garbage": `"garbage" is not a known expand level`,
					},
				},
				"error no ID": {
					expect: map[string]interface{}{
						"+id":  `failed to expand required acting_user: no user ID to expand`,
						"+all": `failed to expand required acting_user: no user ID to expand`,
					},
				},
			},
		},

		{
			name: "channel_member",
			tcs: map[string]TC{
				"happy with API GetChannelMemner": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetChannelMember(channelID, userID).Times(1).Return(&channelMember, nil)
					},
					expect: map[string]interface{}{
						"+summary": expected(apps.ExpandedContext{ChannelMember: &channelMember}),
						"+all":     expected(apps.ExpandedContext{ChannelMember: &channelMember}),
					},
				},
				"happy no API": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expect: map[string]interface{}{
						"+id":  expected(apps.ExpandedContext{ChannelMember: &channelMemberIDOnly}),
						"":     expected(apps.ExpandedContext{}),
						"none": expected(apps.ExpandedContext{}),
					},
				},
				"error no user ID": {
					base: apps.Context{
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expect: map[string]interface{}{
						"+all": "failed to expand required channel_member: no user ID or channel ID to expand",
					},
				},
				"error no channel ID": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[string]interface{}{
						"+all": "failed to expand required channel_member: no user ID or channel ID to expand",
					},
				},
				"error API": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{ChannelID: channelID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetChannelMember(channelID, userID).Times(1).Return(nil, errors.New("ERROR"))
					},
					expect: map[string]interface{}{
						"+all": "failed to expand required channel_member: failed to get channel membership: ERROR",
					},
				},
			},
		},
		{
			name: "team_member",
			tcs: map[string]TC{
				"happy with API GetTeamMember": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expectClientCalls: func(client *mock_mmclient.MockClient) {
						client.EXPECT().GetTeamMember(teamID, userID).Times(1).Return(&teamMember, nil)
					},
					expect: map[string]interface{}{
						"+all":     expected(apps.ExpandedContext{TeamMember: &teamMember}),
						"+summary": expected(apps.ExpandedContext{TeamMember: &teamMember}),
					},
				},
				"happy with no API": {
					base: apps.Context{
						ActingUserID:     userID,
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expect: map[string]interface{}{
						"+id":  expected(apps.ExpandedContext{TeamMember: &teamMemberIDOnly}),
						"":     expected(apps.ExpandedContext{}),
						"none": expected(apps.ExpandedContext{}),
					},
				},
				"no user ID": {
					base: apps.Context{
						UserAgentContext: apps.UserAgentContext{TeamID: teamID},
					},
					expect: map[string]interface{}{
						"+all": "failed to expand required team_member: no user ID or channel ID to expand",
					},
				},
				"no team ID": {
					base: apps.Context{
						ActingUserID: userID,
					},
					expect: map[string]interface{}{
						"+all": "failed to expand required team_member: no user ID or channel ID to expand",
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
					expect: map[string]interface{}{
						"+all": "failed to expand required team_member: failed to get team membership: ERROR",
					},
				},
			},
		},
	} {
		t.Run(field.name, func(t *testing.T) {
			for name, tc := range field.tcs {
				t.Run(name, func(t *testing.T) {
					conf := config.NewTestConfigService(&config.Config{
						MattermostSiteURL: "https://test.mattermost.test",
						DeveloperMode:     true,
					}).WithMattermostConfig(model.Config{
						ServiceSettings: model.ServiceSettings{
							SiteURL: model.NewString("https://test.mattermost.test"),
						},
					})
					for level, expected := range tc.expect {
						t.Run(level, func(t *testing.T) {
							ctrl := gomock.NewController(t)
							client := mock_mmclient.NewMockClient(ctrl)
							if tc.expectClientCalls != nil {
								tc.expectClientCalls(client)
							}
							p := &Proxy{
								conf:                 conf,
								expandClientOverride: client,
							}

							expandData := fmt.Sprintf(`{"%s":"%s"}`, field.name, level)
							e := apps.Expand{}
							err := json.Unmarshal([]byte(expandData), &e)
							require.NoError(t, err)

							r := incoming.NewRequest(conf.MattermostAPI(), conf, utils.NewTestLogger(), nil)
							prev := tc.base
							cc, err := p.expandContext(r, app, &prev, &e)
							if err != nil {
								require.EqualValues(t, expected, err.Error())
							} else {
								require.EqualValues(t, expected, cc.ExpandedContext)
							}
							require.EqualValues(t, prev, tc.base)
						})
					}
				})
			}
		})
	}
}
