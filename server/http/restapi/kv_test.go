// +build !e2e

package restapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/api/mock_api"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

func TestKV(t *testing.T) {
	ctrl := gomock.NewController(t)
	mocked := mock_api.NewMockAppServices(ctrl)
	conf := configurator.NewTestConfigurator(&api.Config{})
	r := mux.NewRouter()
	Init(r, &api.Service{
		Configurator: conf,
		AppServices:  mocked,
	})

	server := httptest.NewServer(r)
	// server := httptest.NewServer(&HH{})
	defer server.Close()

	itemURL := strings.Join([]string{strings.TrimSuffix(server.URL, "/"), api.APIPath, api.KVPath, "/test-id"}, "")
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
	require.NoError(t, err)
	mocked.EXPECT().KVSet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(botUserID, prefix, id string, ref interface{}) (bool, error) {
			require.Equal(t, "01234567890123456789012345", botUserID)
			require.Equal(t, "", prefix)
			require.Equal(t, "test-id", id)
			require.Equal(t, map[string]interface{}{"test_bool": true, "test_string": "test"}, ref)
			return true, nil
		})
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	req, err = http.NewRequest("GET", itemURL, nil)
	require.NoError(t, err)
	req.Header.Set("Mattermost-User-Id", "01234567890123456789012345")
	require.NoError(t, err)
	mocked.EXPECT().KVGet(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
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
