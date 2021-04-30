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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
)

func TestHandleGetBindingsValidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	router := mux.NewRouter()
	Init(router, mm, conf, proxy, nil)

	expected := &apps.Context{
		UserAgentContext: apps.UserAgentContext{
			PostID:    "some_post_id",
			ChannelID: "some_channel_id",
			TeamID:    "some_team_id",
			UserAgent: "webapp",
		},
	}

	bindings := []*apps.Binding{{Location: apps.LocationCommand}}

	proxy.EXPECT().GetBindings("some_session_id", "some_user_id", expected).Return(bindings, nil)
	conf.EXPECT().GetConfig().Return(config.Config{})

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

	req.Header.Add("Mattermost-User-Id", "some_user_id")
	req.Header.Add("MM_SESSION_ID", "some_session_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotNil(t, b)

	bindingsOut := []*apps.Binding{}
	err = json.Unmarshal(b, &bindingsOut)
	require.NoError(t, err)
	require.Equal(t, bindings, bindingsOut)
}
