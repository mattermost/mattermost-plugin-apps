package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
)

type notifyTestcase struct {
	name string
	subs map[string][]apps.Subscription
	run  func(p *Proxy, upstreams map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API)
}

func sendCallResponse(t *testing.T, path string, cr *apps.CallResponse, up *mock_upstream.MockUpstream) {
	b, _ := json.Marshal(cr)
	reader := ioutil.NopCloser(bytes.NewReader(b))
	up.EXPECT().Roundtrip(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(_ apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
		require.Equal(t, path, creq.Path)
		return reader, nil
	})
}

var app1 = apps.App{
	DeployType:  apps.DeployBuiltin,
	BotUserID:   "bot1",
	BotUsername: "bot1username",
	Manifest: apps.Manifest{
		AppID:       apps.AppID("app1"),
		DisplayName: "App 1",
	},
}

var app2 = apps.App{
	DeployType:  apps.DeployBuiltin,
	BotUserID:   "bot2",
	BotUsername: "bot2username",
	Manifest: apps.Manifest{
		AppID:       apps.AppID("app2"),
		DisplayName: "App 2",
	},
}

func TestNotifyMessageHasBeenPosted(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "post_created no subscriptions",
			subs: map[string][]apps.Subscription{
				"sub.bot_mentioned":                {},
				"sub.post_created.some_channel_id": {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				message := "Hey @bot2username!"
				post := &model.Post{
					Message: message,
				}
				err := p.NotifyMessageHasBeenPosted(post, apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "post_created",
			subs: map[string][]apps.Subscription{
				"sub.bot_mentioned": {},
				"sub.post_created.some_channel_id": {
					{
						AppID:     app1.AppID,
						Subject:   apps.SubjectPostCreated,
						ChannelID: "some_channel_id",
						Call:      apps.NewCall("/notify/post_created"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/post_created", cr, up[app1.AppID])

				message := "Hey @bot2username!"
				post := &model.Post{
					Message: message,
				}
				err := p.NotifyMessageHasBeenPosted(post, apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_mentioned, member of channel",
			subs: map[string][]apps.Subscription{
				"sub.post_created.some_channel_id": {},
				"sub.bot_mentioned": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotMentioned,
						Call:    apps.NewCall("/notify/bot_mention1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotMentioned,
						Call:    apps.NewCall("/notify/bot_mention2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/bot_mention2", cr, up[app2.AppID])

				message := "Hey @bot2username!"
				post := &model.Post{
					Message: message,
				}
				testAPI.On("HasPermissionToChannel", "bot2", "", model.PERMISSION_READ_CHANNEL).Return(true)

				err := p.NotifyMessageHasBeenPosted(post, apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_mentioned, member of channel",
			subs: map[string][]apps.Subscription{
				"sub.post_created.some_channel_id": {},
				"sub.bot_mentioned": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotMentioned,
						Call:    apps.NewCall("/notify/bot_mention1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotMentioned,
						Call:    apps.NewCall("/notify/bot_mention2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				message := "Hey @bot2username!"
				post := &model.Post{
					Message: message,
				}
				testAPI.On("HasPermissionToChannel", "bot2", "", model.PERMISSION_READ_CHANNEL).Return(false)

				err := p.NotifyMessageHasBeenPosted(post, apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runNotifyTest(t, []apps.App{app1, app2}, tc)
		})
	}
}

func TestUserHasJoinedChannel(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "user_joined_channel no subscriptions",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_channel.some_channel_id": {},
				"sub.bot_joined_channel":                  {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				err := p.NotifyUserHasJoinedChannel(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "user_joined_channel",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_channel.some_channel_id": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectUserJoinedChannel,
						Call:    apps.NewCall("/notify/user_joined_channel"),
					},
				},
				"sub.bot_joined_channel": {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/user_joined_channel", cr, up[app1.AppID])

				err := p.NotifyUserHasJoinedChannel(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_joined_channel",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_channel.some_channel_id": {},
				"sub.bot_joined_channel": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotJoinedChannel,
						Call:    apps.NewCall("/notify/bot_joined_channel1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotJoinedChannel,
						Call:    apps.NewCall("/notify/bot_joined_channel2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/bot_joined_channel1", cr, up[app1.AppID])

				err := p.NotifyUserHasJoinedChannel(apps.Context{
					UserID: app1.BotUserID,
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func TestUserHasLeftChannel(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "user_left_channel no subscriptions",
			subs: map[string][]apps.Subscription{
				"sub.user_left_channel.some_channel_id": {},
				"sub.bot_left_channel":                  {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				err := p.NotifyUserHasLeftChannel(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "user_left_channel",
			subs: map[string][]apps.Subscription{
				"sub.user_left_channel.some_channel_id": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectUserLeftChannel,
						Call:    apps.NewCall("/notify/user_left_channel"),
					},
				},
				"sub.bot_left_channel": {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/user_left_channel", cr, up[app1.AppID])

				err := p.NotifyUserHasLeftChannel(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_left_channel",
			subs: map[string][]apps.Subscription{
				"sub.user_left_channel.some_channel_id": {},
				"sub.bot_left_channel": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotLeftChannel,
						Call:    apps.NewCall("/notify/bot_left_channel1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotLeftChannel,
						Call:    apps.NewCall("/notify/bot_left_channel2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/bot_left_channel1", cr, up[app1.AppID])

				err := p.NotifyUserHasLeftChannel(apps.Context{
					UserID: app1.BotUserID,
					UserAgentContext: apps.UserAgentContext{
						ChannelID: "some_channel_id",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func TestUserHasJoinedTeam(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "user_joined_team no subscriptions",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_team.some_team_id": {},
				"sub.bot_joined_team":               {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				err := p.NotifyUserHasJoinedTeam(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "user_joined_team",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_team.some_team_id": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectUserJoinedTeam,
						Call:    apps.NewCall("/notify/user_joined_team"),
					},
				},
				"sub.bot_joined_team": {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/user_joined_team", cr, up[app1.AppID])

				err := p.NotifyUserHasJoinedTeam(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_joined_team",
			subs: map[string][]apps.Subscription{
				"sub.user_joined_team.some_team_id": {},
				"sub.bot_joined_team": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotJoinedTeam,
						Call:    apps.NewCall("/notify/bot_joined_team1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotJoinedTeam,
						Call:    apps.NewCall("/notify/bot_joined_team2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/bot_joined_team1", cr, up[app1.AppID])

				err := p.NotifyUserHasJoinedTeam(apps.Context{
					UserID: app1.BotUserID,
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func TestUserHasLeftTeam(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "user_left_team no subscriptions",
			subs: map[string][]apps.Subscription{
				"sub.user_left_team.some_team_id": {},
				"sub.bot_left_team":               {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				err := p.NotifyUserHasLeftTeam(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "user_left_team",
			subs: map[string][]apps.Subscription{
				"sub.user_left_team.some_team_id": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectUserLeftChannel,
						Call:    apps.NewCall("/notify/user_left_team"),
					},
				},
				"sub.bot_left_team": {},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/user_left_team", cr, up[app1.AppID])

				err := p.NotifyUserHasLeftTeam(apps.Context{
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "bot_left_team",
			subs: map[string][]apps.Subscription{
				"sub.user_left_team.some_team_id": {},
				"sub.bot_left_team": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectBotLeftTeam,
						Call:    apps.NewCall("/notify/bot_left_team1"),
					},
					{
						AppID:   app2.AppID,
						Subject: apps.SubjectBotLeftTeam,
						Call:    apps.NewCall("/notify/bot_left_team2"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/bot_left_team1", cr, up[app1.AppID])

				err := p.NotifyUserHasLeftTeam(apps.Context{
					UserID: app1.BotUserID,
					UserAgentContext: apps.UserAgentContext{
						TeamID: "some_team_id",
					},
				})
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func TestChannelHasBeenCreated(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "channel_created",
			subs: map[string][]apps.Subscription{
				"sub.channel_created.some_team_id": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectChannelCreated,
						Call:    apps.NewCall("/notify/channel_created"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/channel_created", cr, up[app1.AppID])

				err := p.Notify(
					apps.Context{
						UserAgentContext: apps.UserAgentContext{
							ChannelID: "some_channel_id",
							TeamID:    "some_team_id",
						},
					},
					apps.SubjectChannelCreated)
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func TestUserHasBeenCreated(t *testing.T) {
	for _, tc := range []notifyTestcase{
		{
			name: "user_created",
			subs: map[string][]apps.Subscription{
				"sub.user_created": {
					{
						AppID:   app1.AppID,
						Subject: apps.SubjectUserCreated,
						Call:    apps.NewCall("/notify/user_created"),
					},
				},
			},
			run: func(p *Proxy, up map[apps.AppID]*mock_upstream.MockUpstream, testAPI *plugintest.API) {
				cr := &apps.CallResponse{
					Type: apps.CallResponseTypeOK,
				}
				sendCallResponse(t, "/notify/user_created", cr, up[app1.AppID])

				err := p.Notify(
					apps.Context{
						UserID: "some_user_id",
						UserAgentContext: apps.UserAgentContext{
							ChannelID: "some_channel_id",
							TeamID:    "some_team_id",
						},
					},
					apps.SubjectUserCreated)
				require.NoError(t, err)
			},
		},
	} {
		runNotifyTest(t, []apps.App{app1, app2}, tc)
	}
}

func runNotifyTest(t *testing.T, allApps []apps.App, tc notifyTestcase) {
	ctrl := gomock.NewController(t)

	conf, testAPI := config.NewTestService(&config.Config{
		PluginURL: "https://test.mattermost.com/plugins/com.mattermost.apps",
	})

	conf = conf.WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("https://test.mattermost.com"),
		},
	})

	s, err := store.MakeService(conf, nil, nil)
	require.NoError(t, err)
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	appMap := map[apps.AppID]apps.App{}
	upMap := map[apps.AppID]upstream.Upstream{}
	upMockMap := map[apps.AppID]*mock_upstream.MockUpstream{}
	for i := range allApps {
		app := allApps[i]
		appMap[app.AppID] = app
		appStore.EXPECT().Get(app.AppID).Return(&app, nil).AnyTimes()

		up := mock_upstream.NewMockUpstream(ctrl)
		upMap[app.AppID] = up
		upMockMap[app.AppID] = up
	}

	appStore.EXPECT().AsMap().Return(appMap).AnyTimes()

	p := &Proxy{
		store:            s,
		builtinUpstreams: upMap,
		conf:             conf,
	}

	for name, subs := range tc.subs {
		b, err := json.Marshal(subs)
		require.NoError(t, err)
		testAPI.On("KVGet", name).Return(b, nil)
	}

	tc.run(p, upMockMap, testAPI)
}
