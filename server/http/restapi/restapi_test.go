package restapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/api/mock_api"
	"github.com/mattermost/mattermost-server/v5/api4"
	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

// Using mocks
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

func TestSubscribe(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	th.ServerTestHelper.LoginSystemAdmin()

	subscription := &apps.Subscription{
		AppID:     "test-apiId",
		Subject:   "test-subject",
		ChannelID: th.ServerTestHelper.BasicChannel.Id,
		TeamID:    th.ServerTestHelper.BasicTeam.Id,
	}

	th.TestForSystemAdmin(t, func(t *testing.T, client *mmclient.ClientPP) {
		_, resp := client.Subscribe(subscription)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
	})
}

func TestUnsubscribe(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	th.ServerTestHelper.LoginSystemAdmin()

	subscription := &apps.Subscription{
		AppID:     "test-apiId",
		Subject:   "test-subject",
		ChannelID: th.ServerTestHelper.BasicChannel.Id,
		TeamID:    th.ServerTestHelper.BasicTeam.Id,
	}

	th.TestForSystemAdmin(t, func(t *testing.T, client *mmclient.ClientPP) {
		// subscribe
		_, resp := client.Subscribe(subscription)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)

		// unsubscribe
		_, resp = client.Unsubscribe(subscription)
		api4.CheckOKStatus(t, resp)
		require.Nil(t, resp.Error)
	})
}

// TODO - add tests for non-bots

func TestKVSet(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	th.ServerTestHelper.LoginSystemAdmin()

	bot := th.ServerTestHelper.CreateBotWithSystemAdminClient()
	th.ServerTestHelper.App.AddUserToTeam(th.ServerTestHelper.BasicTeam.Id, bot.UserId, "")

	rtoken, resp := th.ServerTestHelper.SystemAdminClient.CreateUserAccessToken(bot.UserId, "test token")
	api4.CheckNoError(t, resp)
	th.ClientPP.AuthToken = rtoken.Token
	th.ClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType

	id := "testId"
	prefix := "prefix-test"
	in := map[string]interface{}{}
	in["test_bool"] = true
	in["test_string"] = "test"

	// set
	out, resp := th.ClientPP.KVSet(id, prefix, in)
	api4.CheckOKStatus(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, out["changed"], true)
}

func TestKVGet(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	bot := th.ServerTestHelper.CreateBotWithSystemAdminClient()
	th.ServerTestHelper.App.AddUserToTeam(th.ServerTestHelper.BasicTeam.Id, bot.UserId, "")

	rtoken, resp := th.ServerTestHelper.SystemAdminClient.CreateUserAccessToken(bot.UserId, "test token")
	api4.CheckNoError(t, resp)
	th.ClientPP.AuthToken = rtoken.Token
	th.ClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType

	id := "testId"
	prefix := "prefix-test"
	in := map[string]interface{}{}
	in["test_bool"] = true
	in["test_string"] = "test"

	// set
	outSet, resp := th.ClientPP.KVSet(id, prefix, in)
	api4.CheckOKStatus(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, outSet["changed"], true)

	// get
	outGet, resp := th.ClientPP.KVGet(id, prefix)
	api4.CheckOKStatus(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, outGet["test_bool"], true)
	require.Equal(t, outGet["test_string"], "test")
}

func TestKVDelete(t *testing.T) {
	th := SetupPP(t)
	defer th.TearDown()

	bot := th.ServerTestHelper.CreateBotWithSystemAdminClient()
	th.ServerTestHelper.App.AddUserToTeam(th.ServerTestHelper.BasicTeam.Id, bot.UserId, "")

	rtoken, resp := th.ServerTestHelper.SystemAdminClient.CreateUserAccessToken(bot.UserId, "test token")
	api4.CheckNoError(t, resp)
	th.ClientPP.AuthToken = rtoken.Token
	th.ClientPP.AuthType = th.ServerTestHelper.SystemAdminClient.AuthType

	id := "testId"
	prefix := "prefix-test"
	in := map[string]interface{}{}
	in["test_bool"] = true
	in["test_string"] = "test"

	// set
	outSet, resp := th.ClientPP.KVSet(id, prefix, in)
	api4.CheckOKStatus(t, resp)
	require.Nil(t, resp.Error)
	require.Equal(t, outSet["changed"], true)

	// delete
	_, resp = th.ClientPP.KVDelete(id, prefix)
	api4.CheckNoError(t, resp)

}
