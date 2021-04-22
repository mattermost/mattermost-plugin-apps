package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
)

func TestAppMetadataForClient(t *testing.T) {
	testApps := []*apps.App{
		{
			BotUserID:   "botid",
			BotUsername: "botusername",
			Manifest: apps.Manifest{
				AppID:       apps.AppID("app1"),
				AppType:     apps.AppTypeBuiltin,
				DisplayName: "App 1",
			},
		},
	}

	ctrl := gomock.NewController(t)
	p := newTestProxy(testApps, ctrl)
	c := &apps.CallRequest{
		Context: &apps.Context{
			ContextFromUserAgent: apps.ContextFromUserAgent{
				AppID: "app1",
			},
		},
	}

	resp := p.Call("session_id", "acting_user_id", c)
	require.Equal(t, resp.AppMetadata, &apps.AppMetadataForClient{
		BotUserID:   "botid",
		BotUsername: "botusername",
	})
}

func TestCleanUserCallContext(t *testing.T) {
	t.Run("no context params passed", func(t *testing.T) {
		testAPI := &plugintest.API{}
		testAPI.On("LogDebug", mock.Anything).Return(nil)
		mm := pluginapi.NewClient(testAPI)

		p := Proxy{
			mm: mm,
		}

		userID := "some_user_id"
		cc := &apps.Context{
			ContextFromUserAgent: apps.ContextFromUserAgent{},
		}

		out, err := p.CleanUserCallContext(userID, cc)
		require.Error(t, err)
		require.Nil(t, out)
	})

	t.Run("post id provided in context", func(t *testing.T) {
		t.Run("user is a member of the post's channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			postID := "some_post_id"
			channelID := "some_channel_id"
			teamID := "some_team_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					AppID:     "app1",
					UserAgent: "webapp",
					Location:  "/command",
					PostID:    postID,
					ChannelID: "ignored_channel_id",
					TeamID:    "ignored_team_id",
				},
			}

			testAPI.On("GetPost", "some_post_id").Return(&model.Post{
				Id:        postID,
				ChannelId: channelID,
			}, nil)

			testAPI.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
				ChannelId: channelID,
				UserId:    userID,
			}, nil)

			testAPI.On("GetChannel", channelID).Return(&model.Channel{
				Id:     channelID,
				TeamId: teamID,
			}, nil)

			out, err := p.CleanUserCallContext(userID, cc)
			require.NoError(t, err)
			require.NotNil(t, out)
			expected := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					AppID:     "app1",
					UserAgent: "webapp",
					Location:  "/command",
					PostID:    postID,
					ChannelID: channelID,
					TeamID:    teamID,
				},
			}
			require.Equal(t, expected, out)
		})

		t.Run("user is not a member of the post's channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			postID := "some_post_id"
			channelID := "some_channel_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					PostID:    postID,
					ChannelID: "ignored_channel_id",
					TeamID:    "ignored_team_id",
				},
			}

			testAPI.On("GetPost", "some_post_id").Return(&model.Post{
				Id:        postID,
				ChannelId: channelID,
			}, nil)

			testAPI.On("GetChannelMember", channelID, userID).Return(nil, &model.AppError{
				Message: "user is not a member of the specified channel",
			})

			out, err := p.CleanUserCallContext(userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})

	t.Run("channel id provided in context", func(t *testing.T) {
		t.Run("user is a member of the channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			channelID := "some_channel_id"
			teamID := "some_team_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					ChannelID: channelID,
					TeamID:    "ignored_team_id",
				},
			}

			testAPI.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
				ChannelId: channelID,
				UserId:    userID,
			}, nil)

			testAPI.On("GetChannel", channelID).Return(&model.Channel{
				Id:     channelID,
				TeamId: teamID,
			}, nil)

			out, err := p.CleanUserCallContext(userID, cc)
			require.NoError(t, err)
			require.NotNil(t, out)
			expected := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					ChannelID: channelID,
					TeamID:    teamID,
				},
			}
			require.Equal(t, expected, out)
		})

		t.Run("user is not a member of the channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			channelID := "some_channel_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					ChannelID: channelID,
					TeamID:    "ignored_team_id",
				},
			}

			testAPI.On("GetChannelMember", channelID, userID).Return(nil, &model.AppError{
				Message: "user is not a member of the specified channel",
			})

			out, err := p.CleanUserCallContext(userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})

	t.Run("team id provided in context", func(t *testing.T) {
		t.Run("user is a member of the team", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			teamID := "some_team_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					TeamID: teamID,
				},
			}

			testAPI.On("GetTeamMember", teamID, userID).Return(&model.TeamMember{
				TeamId: teamID,
				UserId: userID,
			}, nil)

			out, err := p.CleanUserCallContext(userID, cc)
			require.NoError(t, err)
			require.NotNil(t, out)
			expected := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					TeamID: teamID,
				},
			}
			require.Equal(t, expected, out)
		})

		t.Run("user is not a member of the team", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			p := Proxy{
				mm: mm,
			}

			userID := "some_user_id"
			teamID := "some_team_id"

			cc := &apps.Context{
				ContextFromUserAgent: apps.ContextFromUserAgent{
					TeamID: teamID,
				},
			}

			testAPI.On("GetTeamMember", teamID, userID).Return(nil, &model.AppError{
				Message: "user is not a member of the specified team",
			})

			out, err := p.CleanUserCallContext(userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})
}

func TestCleanUserCallContextIgnoredValues(t *testing.T) {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	p := Proxy{
		mm: mm,
	}

	userID := "some_user_id"
	postID := "some_post_id"
	channelID := "some_channel_id"
	teamID := "some_team_id"

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			PostID:    postID,
			ChannelID: "ignored_channel_id",
			TeamID:    "ignored_team_id",
		},
		Subject:           "ignored_subject",
		BotUserID:         "ignored_bot_id",
		ActingUserID:      "ignored_acting_user_id",
		UserID:            "ignored_user_id",
		MattermostSiteURL: "ignored_site_url",
		AppPath:           "ignored_app_path",
		ExpandedContext: apps.ExpandedContext{
			BotAccessToken:        "ignored_bot_access_token",
			ActingUser:            &model.User{},
			ActingUserAccessToken: "ignored_user_access_token",
			AdminAccessToken:      "ignored_admin_access_token",
			OAuth2:                apps.OAuth2Context{},
			App:                   &apps.App{},
			Channel:               &model.Channel{},
			Mentioned:             []*model.User{{}},
			Post:                  &model.Post{},
			RootPost:              &model.Post{},
			Team:                  &model.Team{},
			User:                  &model.User{},
		},
	}

	testAPI.On("GetPost", "some_post_id").Return(&model.Post{
		Id:        postID,
		ChannelId: channelID,
	}, nil)

	testAPI.On("GetChannelMember", channelID, userID).Return(&model.ChannelMember{
		ChannelId: channelID,
		UserId:    userID,
	}, nil)

	testAPI.On("GetChannel", channelID).Return(&model.Channel{
		Id:     channelID,
		TeamId: teamID,
	}, nil)

	out, err := p.CleanUserCallContext(userID, cc)
	require.NoError(t, err)
	require.NotNil(t, out)
	expected := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			PostID:    postID,
			ChannelID: channelID,
			TeamID:    teamID,
		},
	}
	require.Equal(t, expected, out)
}

func newTestProxy(testApps []*apps.App, ctrl *gomock.Controller) *Proxy {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	conf := config.NewTestConfigurator(config.Config{}).WithMattermostConfig(model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString("test.mattermost.com"),
		},
	})

	s := store.NewService(mm, conf)
	appStore := mock_store.NewMockAppStore(ctrl)
	s.App = appStore

	upstreams := map[apps.AppID]upstream.Upstream{}
	for _, app := range testApps {
		cr := &apps.CallResponse{
			Type: apps.CallResponseTypeOK,
		}
		b, _ := json.Marshal(cr)
		reader := ioutil.NopCloser(bytes.NewReader(b))

		up := mock_upstream.NewMockUpstream(ctrl)
		up.EXPECT().Roundtrip(gomock.Any(), gomock.Any()).Return(reader, nil)
		upstreams[app.Manifest.AppID] = up
		appStore.EXPECT().Get(app.AppID).Return(app, nil)
	}

	p := &Proxy{
		mm:               mm,
		store:            s,
		builtinUpstreams: upstreams,
		conf:             conf,
	}

	return p
}
