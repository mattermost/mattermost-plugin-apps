package store

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type FakeEncrypter struct{}

func (*FakeEncrypter) Encrypt(message []byte) ([]byte, error) {
	return []byte("qrZ7JgEW2hi37toQsTorIZSqLv4xRDyHfQulLziP3UonAP77idbimFk9dRObgDgOlJj8E9rrFna0ESpSFFj4UQ=="), nil
}

func (*FakeEncrypter) Decrypt(message []byte) ([]byte, error) {
	return []byte(`{"Test1":"test-1","Test2":"test-2"}`), nil
}

func TestCreateOAuth2State(t *testing.T) {
	stateRE := `[A-Za-z0-9-_]+\.[A-Za-z0-9]`
	userID := `userid-test`
	conf, api := config.NewTestService(nil)
	s := oauth2Store{
		Service: &Service{
			conf: conf,
		},
		encrypter: &AESEncrypter{key: "6368616e676520746869732070617373"},
	}

	// CreateState
	api.On("KVSetWithOptions", mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
		require.Regexp(t, KVOAuth2StatePrefix+stateRE, args.String(0))
		data, _ := args.Get(1).([]byte)
		require.Regexp(t, stateRE, string(data))
	}).Return(true, nil)
	urlState, err := s.CreateState(userID)
	require.NoError(t, err)
	key := KVOAuth2StatePrefix + urlState
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
	mismatchedKey := KVOAuth2StatePrefix + mismatchedState
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
	userID := "userIDis26bytes12345678910"
	conf, api := config.NewTestService(nil)

	s := oauth2Store{
		Service: &Service{
			conf: conf,
		},
		encrypter: &FakeEncrypter{},
	}

	type Entity struct {
		Test1, Test2 string
	}
	entity := Entity{"test-1", "test-2"}
	key := ".usome_app_id                     userIDis26bytes12345678910  nYmK(/C@:ZHulkHPF_PY"
	data := []byte(`{"Test1":"test-1","Test2":"test-2"}`)
	dataEncrypted, err := s.encrypter.Encrypt(data)
	assert.NoError(t, err)

	// CreateState
	api.On("KVSetWithOptions", key, dataEncrypted, mock.Anything).Return(true, nil).Once()
	err = s.SaveUser("some_app_id", userID, data)

	assert.NoError(t, err)

	api.On("KVGet", key).Return(dataEncrypted, nil).Once()

	rData, err := s.GetUser("some_app_id", userID)
	assert.NoError(t, err)
	assert.NotNil(t, rData)
	var r Entity
	err = json.Unmarshal(rData, &r)
	assert.NoError(t, err)
	require.Equal(t, entity, r)
}

func TestOAuth2UserNotEncryptedData(t *testing.T) {
	userID := "userIDis26bytes12345678910"
	conf, api := config.NewTestService(nil)

	s := oauth2Store{
		Service: &Service{
			conf: conf,
		},
		encrypter: &FakeEncrypter{},
	}

	type Entity struct {
		Test1, Test2 string
	}
	entity := Entity{"test-1", "test-2"}
	key := ".usome_app_id                     userIDis26bytes12345678910  nYmK(/C@:ZHulkHPF_PY"
	data := []byte(`{"Test1":"test-1","Test2":"test-2"}`)

	// CreateState
	api.On("KVSetWithOptions", key, data, mock.Anything).Return(true, nil).Once()

	// Avoiding using save user so the data is not encrypted when saving
	userkey, err := Hashkey(KVUserPrefix, "some_app_id", userID, "", KVUserKey)
	assert.NoError(t, err)
	_, err = s.conf.MattermostAPI().KV.Set(userkey, data)

	assert.NoError(t, err)

	api.On("KVGet", key).Return(data, nil).Once()

	rData, err := s.GetUser("some_app_id", userID)
	assert.NoError(t, err)
	assert.NotNil(t, rData)
	var r Entity
	err = json.Unmarshal(rData, &r)
	assert.NoError(t, err)
	require.Equal(t, entity, r)
}
