package store

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/configurator"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func TestDeleteSub(t *testing.T) {
	botID := "bot-id"
	mockAPI := &plugintest.API{}
	defer mockAPI.AssertExpectations(t)

	apiClient := pluginapi.NewClient(mockAPI)
	conf := configurator.NewConfigurator(apiClient, nil, &api.BuildConfig{}, botID)
	s := NewStore(apiClient, conf)

	toDelete := api.Subscription{
		Subject:   "user_joined_channel",
		ChannelID: "channel-id",
		AppID:     "app-id",
	}

	storedSubs := []*api.Subscription{
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test1",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test2",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test3",
		},
	}
	storedSubsBytes, _ := json.Marshal(storedSubs)

	storedSubsWithToDelete := []*api.Subscription{
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test1",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test2",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "app-id",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test3",
		},
	}
	storedSubsWithToDeleteBytes, _ := json.Marshal(storedSubsWithToDelete)

	emptySubs := []*api.Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	subKey := "sub_user_joined_channel_channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.DeleteSub(&toDelete)
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, nil).Times(1)
		err := s.DeleteSub(&toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		err := s.DeleteSub(&toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("error setting subscription", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsWithToDeleteBytes, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, storedSubsBytes, mock.Anything).Return(false, model.NewAppError("KVSet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.DeleteSub(&toDelete)
		require.Error(t, err)
		require.Equal(t, "failed to save subscriptions: KVSet: test, test error", err.Error())
	})

	t.Run("subscription not found", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		err := s.DeleteSub(&toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("subscription deleted", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsWithToDeleteBytes, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, storedSubsBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.DeleteSub(&toDelete)
		require.NoError(t, err)
	})
}

func TestGetSubs(t *testing.T) {
	botID := "bot-id"
	mockAPI := &plugintest.API{}
	defer mockAPI.AssertExpectations(t)

	apiClient := pluginapi.NewClient(mockAPI)
	conf := configurator.NewConfigurator(apiClient, nil, &api.BuildConfig{}, botID)
	s := NewStore(apiClient, conf)

	emptySubs := []*api.Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	storedSubs := []*api.Subscription{
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test1",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test2",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test3",
		},
	}

	storedSubsBytes, _ := json.Marshal(storedSubs)

	subKey := "sub_user_joined_channel_channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		_, err := s.LoadSubs("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, nil).Times(1)
		_, err := s.LoadSubs("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		_, err := s.LoadSubs("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("subscription list got", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		subs, err := s.LoadSubs("user_joined_channel", "team-id", "channel-id")
		require.NoError(t, err)
		require.Equal(t, storedSubs, subs)
	})
}

func TestStoreSub(t *testing.T) {
	botID := "bot-id"
	mockAPI := &plugintest.API{}
	defer mockAPI.AssertExpectations(t)

	apiClient := pluginapi.NewClient(mockAPI)
	conf := configurator.NewConfigurator(apiClient, nil, &api.BuildConfig{}, botID)
	s := NewStore(apiClient, conf)

	toStore := api.Subscription{
		Subject:   "user_joined_channel",
		ChannelID: "channel-id",
		AppID:     "app-id",
	}

	storedSubs := []*api.Subscription{
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test1",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test2",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test3",
		},
	}

	storedSubsBytes, _ := json.Marshal(storedSubs)

	storedSubsWithToStore := []*api.Subscription{
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test1",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test2",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "test3",
		},
		{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
			AppID:     "app-id",
		},
	}
	storedSubsWithToStoreBytes, _ := json.Marshal(storedSubsWithToStore)

	emptySubs := []*api.Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	emptySubsWithToStore := []*api.Subscription{&toStore}
	emptySubsWithToStoreBytes, _ := json.Marshal(emptySubsWithToStore)

	subKey := "sub_user_joined_channel_channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.StoreSub(&toStore)
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(nil, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, emptySubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.StoreSub(&toStore)
		require.NoError(t, err)
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, emptySubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.StoreSub(&toStore)
		require.NoError(t, err)
	})

	t.Run("error setting subscription", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, storedSubsWithToStoreBytes, mock.Anything).Return(false, model.NewAppError("KVSet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.StoreSub(&toStore)
		require.Error(t, err)
		require.Equal(t, "KVSet: test, test error", err.Error())
	})

	t.Run("subscription stored", func(t *testing.T) {
		mockAPI.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		mockAPI.On("KVSetWithOptions", subKey, storedSubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.StoreSub(&toStore)
		require.NoError(t, err)
	})
}

func TestSubsKey(t *testing.T) {
	for name, testcase := range map[string]struct {
		Subject   api.Subject
		TeamID    string
		ChannelID string
		Expected  string
	}{
		string(api.SubjectUserCreated): {
			api.SubjectUserCreated,
			"team-id",
			"channel-id",
			"sub_user_created",
		},
		string(api.SubjectUserJoinedChannel): {
			api.SubjectUserJoinedChannel,
			"team-id",
			"channel-id",
			"sub_user_joined_channel_channel-id",
		},
		string(api.SubjectUserLeftChannel): {
			api.SubjectUserLeftChannel,
			"team-id",
			"channel-id",
			"sub_user_left_channel_channel-id",
		},
		string(api.SubjectUserJoinedTeam): {
			api.SubjectUserJoinedTeam,
			"team-id",
			"channel-id",
			"sub_user_joined_team_team-id",
		},
		string(api.SubjectUserLeftTeam): {
			api.SubjectUserLeftTeam,
			"team-id",
			"channel-id",
			"sub_user_left_team_team-id",
		},
		string(api.SubjectUserUpdated): {
			api.SubjectUserUpdated,
			"team-id",
			"channel-id",
			"sub_user_updated",
		},
		string(api.SubjectChannelCreated): {
			api.SubjectChannelCreated,
			"team-id",
			"channel-id",
			"sub_channel_created_team-id",
		},
		string(api.SubjectPostCreated): {
			api.SubjectPostCreated,
			"team-id",
			"channel-id",
			"sub_post_created_channel-id",
		},
	} {
		t.Run(name, func(t *testing.T) {
			r := subsKey(testcase.Subject, testcase.TeamID, testcase.ChannelID)
			require.Equal(t, testcase.Expected, r)
		})
	}
}
