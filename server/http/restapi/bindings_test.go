package restapi

import (
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

func TestHandleGetBindingsInvalidContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	proxy := mock_proxy.NewMockService(ctrl)
	conf := mock_config.NewMockService(ctrl)

	testAPI := &plugintest.API{}
	testAPI.On("LogDebug", mock.Anything).Return(nil)
	mm := pluginapi.NewClient(testAPI)

	router := mux.NewRouter()
	Init(router, mm, conf, proxy, nil)

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(nil, &model.AppError{
		Message: "user is not a member of the specified team",
	})

	recorder := httptest.NewRecorder()
	u := "/api/v1/bindings?team_id=some_team_id"
	req, err := http.NewRequest("GET", u, nil)
	require.NoError(t, err)

	req.Header.Add("Mattermost-User-Id", "some_user_id")
	req.Header.Add("MM_SESSION_ID", "some_session_id")
	router.ServeHTTP(recorder, req)

	resp := recorder.Result()
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotNil(t, b)
	require.Contains(t, string(b), "user is not a member of the specified team")
}

func TestHandleGetBindingsValidContext(t *testing.T) {
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

	bindings := []*apps.Binding{{Location: apps.LocationCommand}}

	testAPI.On("GetTeamMember", "some_team_id", "some_user_id").Return(&model.TeamMember{
		TeamId: "some_team_id",
		UserId: "some_user_id",
	}, nil)

	proxy.EXPECT().GetBindings("some_session_id", "some_user_id", cc).Return(bindings, nil)
	conf.EXPECT().GetConfig().Return(config.Config{})

	recorder := httptest.NewRecorder()
	u := "/api/v1/bindings?team_id=some_team_id"
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
