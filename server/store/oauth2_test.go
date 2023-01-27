package store

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func TestCreateOAuth2State(t *testing.T) {
	stateRE := `[A-Za-z0-9-_]+\.[A-Za-z0-9]`
	userID := `userid-test`
	conf, api := config.NewTestService(nil)
	s := OAuth2Store{}

	// CreateState
	api.On("KVSetWithOptions", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
		require.Regexp(t, KVOAuth2StatePrefix+stateRE, args.String(0))
		data, _ := args.Get(1).([]byte)
		require.Regexp(t, stateRE, string(data))
	}).Return(true, nil)

	r := incoming.NewRequest(conf, nil, "reqid").WithActingUserID(userID).WithSourceAppID("some_app_id")

	urlState, err := s.CreateState(r)
	require.NoError(t, err)
	key := KVOAuth2StatePrefix + urlState
	require.LessOrEqual(t, len(key), model.KeyValueKeyMaxRunes)
	require.Regexp(t, stateRE, urlState)

	// Validate errors
	err = s.ValidateStateOnce(r, "invalidformat")
	require.EqualError(t, err, "forbidden")

	err = s.ValidateStateOnce(r, "nonexistent.value")
	require.EqualError(t, err, "forbidden")

	err = s.ValidateStateOnce(r.WithActingUserID("idmismatch"), urlState)
	require.EqualError(t, err, "forbidden")

	mismatchedState := "mismatched-random." + strings.Split(urlState, ".")[1]
	mismatchedKey := KVOAuth2StatePrefix + mismatchedState
	api.On("KVGet", mismatchedKey).Once().Return(nil, nil)                                          // not found
	api.On("KVSetWithOptions", mismatchedKey, []byte(nil), mock.Anything).Once().Return(false, nil) // delete attempt
	err = s.ValidateStateOnce(r, mismatchedState)
	require.EqualError(t, err, "state mismatch: forbidden")

	api.On("KVGet", key).Once().Return([]byte(`"`+urlState+`"`), nil)
	api.On("KVSetWithOptions", key, []byte(nil), mock.Anything).Once().Return(true, nil) // delete
	err = s.ValidateStateOnce(r, urlState)
	require.NoError(t, err)
}

func TestOAuth2User(t *testing.T) {
	userID := "userIDis26bytes12345678910"
	appID := apps.AppID("some_app_id")
	conf, api := config.NewTestService(nil)
	s := OAuth2Store{}
	r := incoming.NewRequest(conf, nil, "reqid").WithActingUserID(userID).WithSourceAppID(appID)

	type Entity struct {
		Test1, Test2 string
	}
	entity := Entity{"test-1", "test-2"}
	key := ".usome_app_id                     userIDis26bytes12345678910  nYmK(/C@:ZHulkHPF_PY"
	data := []byte(`{"Test1":"test-1","Test2":"test-2"}`)
	// CreateState
	api.On("KVSetWithOptions", key, data, mock.Anything).Return(true, nil).Once()
	err := s.SaveUser(r, data)
	require.NoError(t, err)

	api.On("KVGet", key).Return(data, nil).Once()

	rData, err := s.GetUser(r)
	assert.NoError(t, err)
	assert.NotNil(t, rData)
	var received Entity
	err = json.Unmarshal(rData, &received)
	assert.NoError(t, err)
	require.Equal(t, entity, received)
}
