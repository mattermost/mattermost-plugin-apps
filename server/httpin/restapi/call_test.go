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
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func TestCleanUserAgentContext(t *testing.T) {
	t.Run("no context params passed", func(t *testing.T) {
		a := &restapi{
			conf: config.NewTestConfigService(nil),
		}

		userID := "some_user_id"
		cc := apps.Context{
			UserAgentContext: apps.UserAgentContext{},
		}

		_, err := a.cleanUserAgentContext(userID, cc)
		require.Error(t, err)
	})

	t.Run("post id provided in context", func(t *testing.T) {
		t.Run("user is a member of the post's channel", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			postID := "some_post_id"
			channelID := "some_channel_id"
			teamID := "some_team_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
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

			cc, err := a.cleanUserAgentContext(userID, cc)
			require.NoError(t, err)
			expected := apps.Context{
				ActingUserID: "some_user_id",
				UserAgentContext: apps.UserAgentContext{
					AppID:     "app1",
					UserAgent: "webapp",
					Location:  "/command",
					PostID:    postID,
					ChannelID: channelID,
					TeamID:    teamID,
				},
			}
			require.Equal(t, expected, cc)
		})

		t.Run("user is not a member of the post's channel", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			postID := "some_post_id"
			channelID := "some_channel_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
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

			_, err := a.cleanUserAgentContext(userID, cc)
			require.Error(t, err)
		})
	})

	t.Run("channel id provided in context", func(t *testing.T) {
		t.Run("user is a member of the channel", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			channelID := "some_channel_id"
			teamID := "some_team_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
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

			cc, err := a.cleanUserAgentContext(userID, cc)
			require.NoError(t, err)
			expected := apps.Context{
				ActingUserID: "some_user_id",
				UserAgentContext: apps.UserAgentContext{
					ChannelID: channelID,
					TeamID:    teamID,
				},
			}
			require.Equal(t, expected, cc)
		})

		t.Run("user is not a member of the channel", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			channelID := "some_channel_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
					ChannelID: channelID,
					TeamID:    "ignored_team_id",
				},
			}

			testAPI.On("GetChannelMember", channelID, userID).Return(nil, &model.AppError{
				Message: "user is not a member of the specified channel",
			})

			_, err := a.cleanUserAgentContext(userID, cc)
			require.Error(t, err)
		})
	})

	t.Run("team id provided in context", func(t *testing.T) {
		t.Run("user is a member of the team", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			teamID := "some_team_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
					TeamID: teamID,
				},
			}

			testAPI.On("GetTeamMember", teamID, userID).Return(&model.TeamMember{
				TeamId: teamID,
				UserId: userID,
			}, nil)

			cc, err := a.cleanUserAgentContext(userID, cc)
			require.NoError(t, err)
			expected := apps.Context{
				ActingUserID: "some_user_id",
				UserAgentContext: apps.UserAgentContext{
					TeamID: teamID,
				},
			}
			require.Equal(t, expected, cc)
		})

		t.Run("user is not a member of the team", func(t *testing.T) {
			testConfig, testAPI := config.NewTestService(nil)
			a := &restapi{
				conf: testConfig,
			}

			userID := "some_user_id"
			teamID := "some_team_id"

			cc := apps.Context{
				UserAgentContext: apps.UserAgentContext{
					TeamID: teamID,
				},
			}

			testAPI.On("GetTeamMember", teamID, userID).Return(nil, &model.AppError{
				Message: "user is not a member of the specified team",
			})

			_, err := a.cleanUserAgentContext(userID, cc)
			require.Error(t, err)
		})
	})
}

func TestCleanUserAgentContextIgnoredValues(t *testing.T) {
	testConfig, testAPI := config.NewTestService(nil)
	a := &restapi{
		conf: testConfig,
	}

	userID := "some_user_id"
	postID := "some_post_id"
	channelID := "some_channel_id"
	teamID := "some_team_id"

	cc := apps.Context{
		UserAgentContext: apps.UserAgentContext{
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

	cc, err := a.cleanUserAgentContext(userID, cc)
	require.NoError(t, err)
	expected := apps.Context{
		ActingUserID: "some_user_id",
		UserAgentContext: apps.UserAgentContext{
			PostID:    postID,
			ChannelID: channelID,
			TeamID:    teamID,
		},
	}
	require.Equal(t, expected, cc)
}

func TestHandleCallInvalidContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	p := mock_proxy.NewMockService(ctrl)
	testConfig, testAPI := config.NewTestService(nil)
	router := mux.NewRouter()
	Init(router, testConfig, p, nil)

	call := apps.CallRequest{
		Context: apps.Context{
			UserAgentContext: apps.UserAgentContext{
				TeamID: "some_team_id",
			},
		},
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

	req.Header.Add(config.MattermostUserIDHeader, "some_user_id")
	req.Header.Add(config.MattermostSessionIDHeader, "some_session_id")
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
	p := mock_proxy.NewMockService(ctrl)
	testConfig, testAPI := config.NewTestService(nil)

	router := mux.NewRouter()
	Init(router, testConfig, p, nil)

	creq := apps.CallRequest{
		Call: *apps.NewCall("/path/submit"),
		Context: apps.Context{
			UserAgentContext: apps.UserAgentContext{
				AppID:  "app1",
				TeamID: "some_team_id",
			},
		},
	}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(&model.TeamMember{
		TeamId: "some_team_id",
		UserId: "some_user_id",
	}, nil)

	expected := apps.CallRequest{
		Call: *apps.NewCall("/path/submit"),
		Context: apps.Context{
			ActingUserID: "some_user_id",
			UserAgentContext: apps.UserAgentContext{
				AppID:  "app1",
				TeamID: "some_team_id",
			},
		},
	}

	p.EXPECT().Call(gomock.Any(), expected).Return(proxy.CallResponse{
		CallResponse: apps.NewDataResponse(nil),
	})

	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(creq)
	require.NoError(t, err)

	u := "/api/v1/call"
	req, err := http.NewRequest("POST", u, b)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()

	req.Header.Add(config.MattermostUserIDHeader, "some_user_id")
	req.Header.Add(config.MattermostSessionIDHeader, "some_session_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
