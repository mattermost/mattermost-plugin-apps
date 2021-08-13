// +build !e2e

package restapi

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mocks/mock_store"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func TestKV(t *testing.T) {
	testConfig, testAPI := config.NewTestService(nil)
	testAPI.On("GetUser", mock.Anything).Return(
		&model.User{
			IsBot: true,
		}, nil)

	ctrl := gomock.NewController(t)
	mocked := mock_store.NewMockAppKVStore(ctrl)
	mockStore := &store.Service{
		AppKV: mocked,
	}

	appService := appservices.NewService(testConfig, mockStore)

	r := mux.NewRouter()
	Init(r, testConfig, nil, appService)

	server := httptest.NewServer(r)
	defer server.Close()

	itemURL := strings.Join([]string{strings.TrimSuffix(server.URL, "/"), mmclient.PathAPI, mmclient.PathKV, "/test-id"}, "")
	item := []byte(`{"test_string":"test","test_bool":true}`)

	req, err := http.NewRequest("PUT", itemURL, bytes.NewReader(item))
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("PUT", itemURL, bytes.NewReader(item))
	require.NoError(t, err)
	req.Header.Set("Mattermost-User-Id", "01234567890123456789012345")
	req.Header.Set("Mattermost-Session-Id", "01234567890123456789012345")
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

	req, err = http.NewRequest("GET", itemURL, nil)
	require.NoError(t, err)
	req.Header.Set("Mattermost-User-Id", "01234567890123456789012345")
	req.Header.Set("Mattermost-Session-Id", "01234567890123456789012345")
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
