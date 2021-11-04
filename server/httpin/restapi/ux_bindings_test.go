package restapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_session"
)

func TestHandleGetBindingsValidContext(t *testing.T) {
	conf := config.NewTestConfigService(nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	proxy := mock_proxy.NewMockService(ctrl)
	appServices := mock_appservices.NewMockService(ctrl)
	sessionService := mock_session.NewMockService(ctrl)

	router := mux.NewRouter()
	Init(router, conf, proxy, appServices, sessionService)

	expected := apps.Context{
		UserAgentContext: apps.UserAgentContext{
			PostID:    "some_post_id",
			ChannelID: "some_channel_id",
			TeamID:    "some_team_id",
			UserAgent: "webapp",
		},
	}

	bindings := []apps.Binding{{Location: apps.LocationCommand}}

	proxy.EXPECT().GetBindings(gomock.Any(), expected).Return(bindings, nil)

	query := url.Values{
		"post_id":         {"some_post_id"},
		"channel_id":      {"some_channel_id"},
		"team_id":         {"some_team_id"},
		"user_agent_type": {"webapp"},
	}
	q := query.Encode()

	recorder := httptest.NewRecorder()
	u := "/api/v1/bindings?" + q
	req, err := http.NewRequest("GET", u, nil)
	require.NoError(t, err)

	req.Header.Add(config.MattermostUserIDHeader, "some_user_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotNil(t, b)

	bindingsOut := []apps.Binding{}
	err = json.Unmarshal(b, &bindingsOut)
	require.NoError(t, err)
	require.Equal(t, bindings, bindingsOut)
}
