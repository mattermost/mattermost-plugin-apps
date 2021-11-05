package restapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_session"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func TestKV(t *testing.T) {
	conf, testAPI := config.NewTestService(nil)
	testAPI.On("GetUser", mock.Anything).Return(
		&model.User{
			IsBot: true,
		}, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mocked := mock_store.NewMockAppKVStore(ctrl)
	mockStore := &store.Service{
		AppKV: mocked,
	}
	proxy := mock_proxy.NewMockService(ctrl)
	appService := appservices.NewService(conf, mockStore)
	sessionService := mock_session.NewMockService(ctrl)

	router := mux.NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()
	Init(request.NewContext(nil, conf, sessionService), router, proxy, appService)

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
	require.NoError(t, err)
	mocked.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(botUserID, prefix, id string, ref interface{}) (bool, error) {
			require.Equal(t, "01234567890123456789012345", botUserID)
			require.Equal(t, "", prefix)
			require.Equal(t, "test-id", id)
			require.Equal(t, `{"test_string":"test","test_bool":true}`, fmt.Sprintf("%s", ref))
			return true, nil
		})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest(http.MethodGet, itemURL, nil)
	require.NoError(t, err)
	req.Header.Set(config.MattermostUserIDHeader, "01234567890123456789012345")
	require.NoError(t, err)
	mocked.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(botUserID, prefix, id string, ref interface{}) (bool, error) {
			require.Equal(t, "01234567890123456789012345", botUserID)
			require.Equal(t, "", prefix)
			require.Equal(t, "test-id", id)
			return true, nil
		})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestKVPut(t *testing.T) {
	t.Run("payload too big", func(t *testing.T) {
		conf := config.NewTestConfigService(nil)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		proxy := mock_proxy.NewMockService(ctrl)
		appServices := mock_appservices.NewMockService(ctrl)
		sessionService := mock_session.NewMockService(ctrl)

		router := mux.NewRouter()
		server := httptest.NewServer(router)
		defer server.Close()
		Init(request.NewContext(nil, conf, sessionService), router, proxy, appServices)

		payload := make([]byte, MaxKVStoreValueLength+1)
		expectedPayload := make([]byte, MaxKVStoreValueLength)

		appServices.EXPECT().KVSet("some_user_id", "", "some_key", expectedPayload).Return(true, nil)

		u := server.URL + path.API + path.KV + "/some_key"
		body := bytes.NewReader(payload)
		req, err := http.NewRequest(http.MethodPut, u, body)
		require.NoError(t, err)
		req.Header.Add(config.MattermostUserIDHeader, "some_user_id")

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
