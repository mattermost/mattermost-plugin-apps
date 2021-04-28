package restapi

import (
	"bytes"
	"encoding/json"
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

func TestCleanUserCallContextIgnoredValues(t *testing.T) {
	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	a := &restapi{
		mm: mm,
	}

	cc := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			PostID:    "some_post_id",
			ChannelID: "some_channel_id",
			TeamID:    "some_team_id",
			UserAgent: "webapp",
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

	out := cleanUserCallContext(a.mm, "some_user_id", cc)
	require.NotNil(t, out)
	expected := &apps.Context{
		ContextFromUserAgent: apps.ContextFromUserAgent{
			PostID:    "some_post_id",
			ChannelID: "some_channel_id",
			TeamID:    "some_team_id",
			UserAgent: "webapp",
		},
	}
	require.Equal(t, expected, out)
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
			PostID:    "some_post_id",
			ChannelID: "some_channel_id",
			TeamID:    "some_team_id",
			UserAgent: "webapp",
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
	call := &apps.CallRequest{
		Context: cc,
	}

	expected := &apps.CallRequest{
		Context: &apps.Context{
			ContextFromUserAgent: apps.ContextFromUserAgent{
				PostID:    "some_post_id",
				ChannelID: "some_channel_id",
				TeamID:    "some_team_id",
				UserAgent: "webapp",
			},
		},
	}

	proxy.EXPECT().Call("some_session_id", "some_user_id", expected).Return(&apps.ProxyCallResponse{})

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
