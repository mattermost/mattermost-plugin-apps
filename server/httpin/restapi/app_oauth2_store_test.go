package restapi

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_proxy"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_session"
)

func TestOAuth2StoreUser(t *testing.T) {
	t.Run("small payload", func(t *testing.T) {
		conf := config.NewTestConfigService(nil)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		proxy := mock_proxy.NewMockService(ctrl)
		appServices := mock_appservices.NewMockService(ctrl)
		sessionService := mock_session.NewMockService(ctrl)

		router := mux.NewRouter()
		server := httptest.NewServer(router)
		defer server.Close()
		Init(router, conf, proxy, appServices, sessionService)

		payload := []byte("some payload")
		expectedPayload := payload
		appServices.EXPECT().StoreOAuth2User(apps.AppID("some_app_id"), "some_user_id", expectedPayload).Return(nil)

		u := server.URL + path.API + path.OAuth2User + "/some_app_id"
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
		assert.Empty(t, b)
	})

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
		Init(router, conf, proxy, appServices, sessionService)

		payload := make([]byte, MaxKVStoreValueLength+1)
		expectedPayload := make([]byte, MaxKVStoreValueLength)

		appServices.EXPECT().StoreOAuth2User(apps.AppID("some_app_id"), "some_user_id", expectedPayload).Return(nil)

		u := server.URL + path.API + path.OAuth2User + "/some_app_id"
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
		assert.Empty(t, b)
	})
}
