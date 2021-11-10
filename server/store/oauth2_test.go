package store

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func TestCreateOAuth2State(t *testing.T) {
	stateRE := `[A-Za-z0-9-_]+\.[A-Za-z0-9]`
	userID := `userid-test`
	conf, api := config.NewTestService(nil)
	s := oauth2Store{
		Service: &Service{
			conf: conf,
		},
	}

	// CreateState
	api.On("KVSetWithOptions", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
		require.Regexp(t, config.KVOAuth2StatePrefix+stateRE, args.String(0))
		data, _ := args.Get(1).([]byte)
		require.Regexp(t, stateRE, string(data))
	}).Return(true, nil)
	urlState, err := s.CreateState(userID)
	require.NoError(t, err)
	key := config.KVOAuth2StatePrefix + urlState
	require.LessOrEqual(t, len(key), model.KeyValueKeyMaxRunes)
	require.Regexp(t, stateRE, urlState)

	// Validate errors
	err = s.ValidateStateOnce("invalidformat", userID)
	require.EqualError(t, err, "forbidden")

	err = s.ValidateStateOnce("nonexistent.value", userID)
	require.EqualError(t, err, "forbidden")

	err = s.ValidateStateOnce(urlState, "idmismatch")
	require.EqualError(t, err, "forbidden")

	mismatchedState := "mismatched-random." + strings.Split(urlState, ".")[1]
	mismatchedKey := config.KVOAuth2StatePrefix + mismatchedState
	api.On("KVGet", mismatchedKey).Once().Return(nil, nil)                                          // not found
	api.On("KVSetWithOptions", mismatchedKey, []byte(nil), mock.Anything).Once().Return(false, nil) // delete attempt
	err = s.ValidateStateOnce(mismatchedState, userID)
	require.EqualError(t, err, "state mismatch: forbidden")

	api.On("KVGet", key).Once().Return([]byte(`"`+urlState+`"`), nil)
	api.On("KVSetWithOptions", key, []byte(nil), mock.Anything).Once().Return(true, nil) // delete
	err = s.ValidateStateOnce(urlState, userID)
	require.NoError(t, err)
}

func TestOAuth2User(t *testing.T) {
	userID := `userid-test`
	conf, api := config.NewTestService(nil)
	s := oauth2Store{
		Service: &Service{
			conf: conf,
		},
	}

	type Entity struct {
		Test1, Test2 string
	}
	entity := Entity{"test-1", "test-2"}
	key := ".ubotUserIDis26bytes90123456  <B0k.Len6V#gQ?bE*:UQ"
	data := `{"Test1":"test-1","Test2":"test-2"}`
	// CreateState
	api.On("KVSetWithOptions", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
		require.Equal(t, key, args.String(0))
		setData, _ := args.Get(1).([]byte)
		require.Equal(t, data, string(setData))
	}).Return(true, nil)
	err := s.SaveUser("botUserIDis26bytes90123456", userID, &entity)
	require.NoError(t, err)

	api.On("KVGet", key).Once().Return([]byte(data), nil)
	r := Entity{}
	err = s.GetUser("botUserIDis26bytes90123456", userID, &r)
	require.NoError(t, err)
	require.Equal(t, entity, r)
}
