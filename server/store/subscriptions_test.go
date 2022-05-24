package store

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestDeleteSub(t *testing.T) {
	conf, api := config.NewTestService(nil)
	defer api.AssertExpectations(t)

	s, err := MakeService(utils.NewTestLogger(), conf, nil)
	require.NoError(t, err)

	toDelete := Subscription{
		Subscription: apps.Subscription{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
		},
		AppID: "app-id",
	}

	storedSubs := []Subscription{
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test1",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test2",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test3",
		},
	}
	storedSubsBytes, _ := json.Marshal(storedSubs)

	storedSubsWithToDelete := []Subscription{
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			}, AppID: "test1",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			}, AppID: "test2",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "app-id",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test3",
		},
	}
	storedSubsWithToDeleteBytes, _ := json.Marshal(storedSubsWithToDelete)

	emptySubs := []apps.Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	subKey := "sub.user_joined_channel.channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, nil).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("error setting subscription", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsWithToDeleteBytes, nil).Times(1)
		api.On("KVSetWithOptions", subKey, storedSubsBytes, mock.Anything).Return(false, model.NewAppError("KVSet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.Error(t, err)
		require.Equal(t, "failed to save subscriptions: KVSet: test, test error", err.Error())
	})

	t.Run("subscription not found", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("subscription deleted", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsWithToDeleteBytes, nil).Times(1)
		api.On("KVSetWithOptions", subKey, storedSubsBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.Subscription.Delete(toDelete)
		require.NoError(t, err)
	})
}

func TestGetSubs(t *testing.T) {
	conf, api := config.NewTestService(nil)
	defer api.AssertExpectations(t)
	s, err := MakeService(utils.NewTestLogger(), conf, nil)
	require.NoError(t, err)

	emptySubs := []apps.Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	storedSubs := []Subscription{
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test1",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test2",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test3",
		},
	}

	storedSubsBytes, _ := json.Marshal(storedSubs)

	subKey := "sub.user_joined_channel.channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		_, err := s.Subscription.Get("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, nil).Times(1)
		_, err := s.Subscription.Get("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		_, err := s.Subscription.Get("user_joined_channel", "team-id", "channel-id")
		require.Error(t, err)
		require.Equal(t, utils.ErrNotFound.Error(), err.Error())
	})

	t.Run("subscription list got", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		subs, err := s.Subscription.Get("user_joined_channel", "team-id", "channel-id")
		require.NoError(t, err)
		require.Equal(t, storedSubs, subs)
	})
}

func TestStoreSub(t *testing.T) {
	conf, api := config.NewTestService(nil)
	defer api.AssertExpectations(t)
	s, err := MakeService(utils.NewTestLogger(), conf, nil)
	require.NoError(t, err)

	toStore := Subscription{
		Subscription: apps.Subscription{
			Subject:   "user_joined_channel",
			ChannelID: "channel-id",
		},
		AppID: "app-id",
	}

	storedSubs := []Subscription{
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			}, AppID: "test1",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test2",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test3",
		},
	}

	storedSubsBytes, _ := json.Marshal(storedSubs)

	storedSubsWithToStore := []Subscription{
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			}, AppID: "test1",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			}, AppID: "test2",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "test3",
		},
		{
			Subscription: apps.Subscription{
				Subject:   "user_joined_channel",
				ChannelID: "channel-id",
			},
			AppID: "app-id",
		},
	}
	storedSubsWithToStoreBytes, _ := json.Marshal(storedSubsWithToStore)

	emptySubs := []Subscription{}
	emptySubsBytes, _ := json.Marshal(emptySubs)

	emptySubsWithToStore := []Subscription{toStore}
	emptySubsWithToStoreBytes, _ := json.Marshal(emptySubsWithToStore)

	subKey := "sub.user_joined_channel.channel-id"

	t.Run("error getting subscriptions", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, model.NewAppError("KVGet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.Subscription.Save(toStore)
		require.Error(t, err)
		require.Equal(t, "KVGet: test, test error", err.Error())
	})

	t.Run("no value for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(nil, nil).Times(1)
		api.On("KVSetWithOptions", subKey, emptySubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.Subscription.Save(toStore)
		require.NoError(t, err)
	})

	t.Run("empty list for subs key", func(t *testing.T) {
		api.On("KVGet", subKey).Return(emptySubsBytes, nil).Times(1)
		api.On("KVSetWithOptions", subKey, emptySubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.Subscription.Save(toStore)
		require.NoError(t, err)
	})

	t.Run("error setting subscription", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		api.On("KVSetWithOptions", subKey, storedSubsWithToStoreBytes, mock.Anything).Return(false, model.NewAppError("KVSet", "test", map[string]interface{}{}, "test error", 0)).Times(1)
		err := s.Subscription.Save(toStore)
		require.Error(t, err)
		require.Equal(t, "KVSet: test, test error", err.Error())
	})

	t.Run("subscription stored", func(t *testing.T) {
		api.On("KVGet", subKey).Return(storedSubsBytes, nil).Times(1)
		api.On("KVSetWithOptions", subKey, storedSubsWithToStoreBytes, mock.Anything).Return(true, nil).Times(1)
		err := s.Subscription.Save(toStore)
		require.NoError(t, err)
	})
}

func TestSubsKey(t *testing.T) {
	for name, testcase := range map[string]struct {
		Subject   apps.Subject
		TeamID    string
		ChannelID string
		Expected  string
	}{
		string(apps.SubjectUserCreated): {
			apps.SubjectUserCreated,
			"team-id",
			"channel-id",
			"sub.user_created",
		},
		string(apps.SubjectUserJoinedChannel): {
			apps.SubjectUserJoinedChannel,
			"team-id",
			"channel-id",
			"sub.user_joined_channel.channel-id",
		},
		string(apps.SubjectUserLeftChannel): {
			apps.SubjectUserLeftChannel,
			"team-id",
			"channel-id",
			"sub.user_left_channel.channel-id",
		},
		string(apps.SubjectUserJoinedTeam): {
			apps.SubjectUserJoinedTeam,
			"team-id",
			"channel-id",
			"sub.user_joined_team.team-id",
		},
		string(apps.SubjectUserLeftTeam): {
			apps.SubjectUserLeftTeam,
			"team-id",
			"channel-id",
			"sub.user_left_team.team-id",
		},
		string(apps.SubjectChannelCreated): {
			apps.SubjectChannelCreated,
			"team-id",
			"channel-id",
			"sub.channel_created.team-id",
		},
		// string(apps.SubjectPostCreated): {
		// 	apps.SubjectPostCreated,
		// 	"team-id",
		// 	"channel-id",
		// 	"sub.post_created.channel-id",
		// },
	} {
		t.Run(name, func(t *testing.T) {
			r, err := subsKey(testcase.Subject, testcase.TeamID, testcase.ChannelID)
			require.NoError(t, err)
			require.Equal(t, testcase.Expected, r)
		})
	}
}
