package restapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestKV(t *testing.T) {
	conf, api := config.NewTestService(nil)
	defer api.AssertExpectations(t)
	session := &model.Session{}
	session.AddProp(model.SessionPropMattermostAppID, "some_app_id")
	api.On("GetSession", "some_session_id").Return(session, nil)

	ctrl := gomock.NewController(t)

	mocked := mock_store.NewMockAppKVStore(ctrl)
	mockStore := &store.Service{
		AppKV: mocked,
	}
	proxy := mock_proxy.NewMockService(ctrl)
	appService := appservices.NewService(conf, mockStore)
	// TODO: <>/<> impossible, but is it necessary?
	// proxy.sessionService = mock_session.NewMockService(ctrl)

	router := mux.NewRouter()
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	rh := handler.NewHandler(proxy, conf, utils.NewTestLogger())
	Init(rh, appService)

	itemURL := strings.Join([]string{strings.TrimSuffix(server.URL, "/"), path.API, path.KV, "/test-id"}, "")
	item := []byte(`{"test_string":"test","test_bool":true}`)

	req, err := http.NewRequest(http.MethodPut, itemURL, bytes.NewReader(item))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest(http.MethodPut, itemURL, bytes.NewReader(item))
	require.NoError(t, err)
	req.Header.Set(config.MattermostUserIDHeader, "01234567890123456789012345")
	req.Header.Add(config.MattermostSessionIDHeader, "some_session_id")
	require.NoError(t, err)
	mocked.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(r *incoming.Request, prefix, id string, ref []byte) (bool, error) {
			assert.Equal(t, apps.AppID("some_app_id"), r.SourceAppID())
			assert.Equal(t, "01234567890123456789012345", r.ActingUserID())
			assert.Equal(t, "", prefix)
			assert.Equal(t, "test-id", id)
			assert.Equal(t, []byte(`{"test_string":"test","test_bool":true}`), ref)
			return true, nil
		})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest(http.MethodGet, itemURL, nil)
	require.NoError(t, err)
	req.Header.Set(config.MattermostUserIDHeader, "01234567890123456789012345")
	req.Header.Add(config.MattermostSessionIDHeader, "some_session_id")
	require.NoError(t, err)
	mocked.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(r *incoming.Request, prefix, id string) ([]byte, error) {
			assert.Equal(t, apps.AppID("some_app_id"), r.SourceAppID())
			require.Equal(t, "01234567890123456789012345", r.ActingUserID())
			require.Equal(t, "", prefix)
			require.Equal(t, "test-id", id)
			return item, nil
		})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestKVPut(t *testing.T) {
	t.Run("payload too big", func(t *testing.T) {
		conf, api := config.NewTestService(nil)
		defer api.AssertExpectations(t)
		session := &model.Session{}
		session.AddProp(model.SessionPropMattermostAppID, "some_app_id")
		api.On("GetSession", "some_session_id").Return(session, nil)

		ctrl := gomock.NewController(t)
		proxy := mock_proxy.NewMockService(ctrl)
		appServices := mock_appservices.NewMockService(ctrl)
		// TODO: <>/<> impossible, but is it necessary?
		// sessionService := mock_session.NewMockService(ctrl)

		router := mux.NewRouter()
		server := httptest.NewServer(router)
		t.Cleanup(server.Close)
		rh := handler.NewHandler(proxy, conf, utils.NewTestLogger())
		Init(rh, appServices)

		payload := make([]byte, MaxKVStoreValueLength+1)
		expectedPayload := make([]byte, MaxKVStoreValueLength)

		appServices.EXPECT().KVSet(gomock.Any(), apps.AppID("some_app_id"), "some_user_id", "", "some_key", expectedPayload).Return(true, nil)

		u := server.URL + path.API + path.KV + "/some_key"
		body := bytes.NewReader(payload)
		req, err := http.NewRequest(http.MethodPut, u, body)
		require.NoError(t, err)
		req.Header.Add(config.MattermostUserIDHeader, "some_user_id")
		req.Header.Add(config.MattermostSessionIDHeader, "some_session_id")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		b, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.NoError(t, err)
		assert.NotNil(t, b)
	})
}
