package restapi

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
)

func TestCleanUserCallContext(t *testing.T) {
	t.Run("no context params passed", func(t *testing.T) {
		testAPI := &plugintest.API{}
		testAPI.On("LogDebug", mock.Anything).Return(nil)
		mm := pluginapi.NewClient(testAPI)

		a := &restapi{
			mm: mm,
		}

		userID := "some_user_id"
		cc := &apps.Context{
			ContextFromUserAgent: apps.ContextFromUserAgent{},
		}

		out, err := cleanUserCallContext(a.mm, userID, cc)
		require.Error(t, err)
		require.Nil(t, out)
	})

	t.Run("post id provided in context", func(t *testing.T) {
		t.Run("user is a member of the post's channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
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

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})

	t.Run("channel id provided in context", func(t *testing.T) {
		t.Run("user is a member of the channel", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
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

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})

	t.Run("team id provided in context", func(t *testing.T) {
		t.Run("user is a member of the team", func(t *testing.T) {
			testAPI := &plugintest.API{}
			testAPI.On("LogDebug", mock.Anything).Return(nil)
			mm := pluginapi.NewClient(testAPI)

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
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

			a := &restapi{
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

			out, err := cleanUserCallContext(a.mm, userID, cc)
			require.Error(t, err)
			require.Nil(t, out)
		})
	})
}

func TestCleanUserCallContextIgnoredValues(t *testing.T) {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	a := &restapi{
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

	out, err := cleanUserCallContext(a.mm, userID, cc)
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

func TestHandleCallInvalidContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	router := mux.NewRouter()
	Init(router, mm, conf, proxy, nil)

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}
	call := &apps.CallRequest{
		Context: cc,
	}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(nil, &model.AppError{
		Message: "user is not a member of the specified team",
	})

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(call)
	require.NoError(t, err)

	u := "/api/v1/call"
	req, err := http.NewRequest("POST", u, b)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	req.Header.Add("Mattermost-User-Id", "some_user_id")
	req.Header.Add("MM_SESSION_ID", "some_session_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Contains(t, string(resBody), "user is not a member of the specified team")
}

func TestHandleCallValidContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	router := mux.NewRouter()
	Init(router, mm, conf, proxy, nil)

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			TeamID: "some_team_id",
		},
	}
	call := &apps.CallRequest{
		Context: cc,
	}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(&model.TeamMember{
		TeamId: "some_team_id",
		UserId: "some_user_id",
	}, nil)

	proxy.EXPECT().Call("some_session_id", "some_user_id", call).Return(&apps.ProxyCallResponse{})

	conf.EXPECT().GetConfig().Return(config.Config{})

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(call)
	require.NoError(t, err)

	u := "/api/v1/call"
	req, err := http.NewRequest("POST", u, b)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	req.Header.Add("Mattermost-User-Id", "some_user_id")
	req.Header.Add("MM_SESSION_ID", "some_session_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
