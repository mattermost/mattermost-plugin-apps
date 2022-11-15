package store

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestCachedStore(t *testing.T) {
	// stateRE := `[A-Za-z0-9-_]+\.[A-Za-z0-9]`
	// userID := `userid-test`
	conf, api := config.NewTestService(nil)

	type Test struct {
		ID   string
		Body string
	}

	api.On("KVGet", ".cached.test-index").Once().
		Return([]byte(nil), (*model.AppError)(nil))
	s, err := MakeCachedStore[Test]("test", api, conf.MattermostAPI(), conf.Logger())
	require.NoError(t, err)

	r := incoming.NewRequest(conf, conf.Logger(), nil)
	put := func(id string, data Test, indexBefore, indexAfter string) {
		api.On("KVSetWithOptions", "mutex_.cached.test-mutex", []byte{0x1}, mock.Anything).Once().
			Return(true, nil)
		api.On("KVGet", ".cached.test-index").Once().
			Return([]byte(indexBefore), (*model.AppError)(nil))
		api.On("KVSetWithOptions", fmt.Sprintf(".cached.test-item-%s", id), []byte(utils.ToJSON(data)), mock.Anything).Once().
			Return(true, nil)
		api.On("KVSetWithOptions", ".cached.test-index", []byte(indexAfter), mock.Anything).Once().
			Return(true, nil)

		api.On(
			"PublishPluginClusterEvent",
			mock.Anything,
			model.PluginClusterEventSendOptions{
				SendType: "reliable",
			},
		).
			Once().
			Run(
				func(args mock.Arguments) {
					e, ok := args[0].(model.PluginClusterEvent)
					require.True(t, ok)
					require.Equal(t, CachedStoreEventID, e.Id)
					require.NotEmpty(t, e.Data)

					var event cachedStoreEvent[Test]
					err = json.Unmarshal(e.Data, &event)
					require.NoError(t, err)

					require.Equal(t, CachedStorePutMethod, event.Method)
					require.Equal(t, "test", event.StoreName)
					require.Equal(t, id, event.Key)
					require.Equal(t, data, event.Data)
					require.NotEmpty(t, event.SentAt)

					opts, ok := args[1].(model.PluginClusterEventSendOptions)
					require.True(t, ok)
					require.Equal(t, "reliable", opts.SendType)
				}).
			Return(nil)

		api.On("KVSetWithOptions", "mutex_.cached.test-mutex", []byte(nil), mock.Anything).Once().
			Return(true, nil)

		err = s.Put(r, id, data)
		require.NoError(t, err)
		api.AssertExpectations(t)
	}

	delete := func(id string, indexBefore, indexAfter string) {
		api.On("KVSetWithOptions", "mutex_.cached.test-mutex", []byte{0x1}, mock.Anything).Once().
			Return(true, nil)
		api.On("KVGet", ".cached.test-index").Once().
			Return([]byte(indexBefore), (*model.AppError)(nil))
		api.On("KVSetWithOptions", fmt.Sprintf(".cached.test-item-%s", id), []byte(nil), mock.Anything).Once().
			Return(true, nil)
		api.On("KVSetWithOptions", ".cached.test-index", []byte(indexAfter), mock.Anything).Once().
			Return(true, nil)

		api.On(
			"PublishPluginClusterEvent",
			mock.Anything,
			model.PluginClusterEventSendOptions{
				SendType: "reliable",
			},
		).
			Once().
			Run(
				func(args mock.Arguments) {
					e, ok := args[0].(model.PluginClusterEvent)
					require.True(t, ok)
					require.Equal(t, CachedStoreEventID, e.Id)
					require.NotEmpty(t, e.Data)

					var event cachedStoreEvent[Test]
					err = json.Unmarshal(e.Data, &event)
					require.NoError(t, err)

					require.Equal(t, CachedStoreDeleteMethod, event.Method)
					require.Equal(t, "test", event.StoreName)
					require.Equal(t, id, event.Key)
					require.NotEmpty(t, event.SentAt)

					opts, ok := args[1].(model.PluginClusterEventSendOptions)
					require.True(t, ok)
					require.Equal(t, "reliable", opts.SendType)
				}).
			Return(nil)

		api.On("KVSetWithOptions", "mutex_.cached.test-mutex", []byte(nil), mock.Anything).Once().
			Return(true, nil)

		err = s.Delete(r, id)
		require.NoError(t, err)
		api.AssertExpectations(t)
	}

	t.Run("happy put and delete", func(t *testing.T) {
		put("1",
			Test{ID: "1", Body: "test1"},
			``,
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"}]}`,
		)
		put("2",
			Test{ID: "2", Body: "test2"},
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"}]}`,
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"76e8644d57eabcb1b39ae54908855de7bb6b53426f1519d88e772e16a18112ce"}]}`)

		put("3",
			Test{ID: "3", Body: "test3"},
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"76e8644d57eabcb1b39ae54908855de7bb6b53426f1519d88e772e16a18112ce"}]}`,
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"76e8644d57eabcb1b39ae54908855de7bb6b53426f1519d88e772e16a18112ce"},{"k":"3","h":"d0b3bd259bb16568511cc24e41411895c16a58a579f00d9c8638b724c39824d0"}]}`)

		put("2",
			Test{ID: "2", Body: "test2-updated"},
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"76e8644d57eabcb1b39ae54908855de7bb6b53426f1519d88e772e16a18112ce"},{"k":"3","h":"d0b3bd259bb16568511cc24e41411895c16a58a579f00d9c8638b724c39824d0"}]}`,
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"8e72172076018bcaf12a91703674fc1fc9ab907d0f805a89d52f513bf95fca00"},{"k":"3","h":"d0b3bd259bb16568511cc24e41411895c16a58a579f00d9c8638b724c39824d0"}]}`)

		delete("2",
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"2","h":"8e72172076018bcaf12a91703674fc1fc9ab907d0f805a89d52f513bf95fca00"},{"k":"3","h":"d0b3bd259bb16568511cc24e41411895c16a58a579f00d9c8638b724c39824d0"}]}`,
			`{"Data":[{"k":"1","h":"a74d512a5bc25ef815367c51c2bfa7535d1d73079fd9909f4a7d2ef4d256ff22"},{"k":"3","h":"d0b3bd259bb16568511cc24e41411895c16a58a579f00d9c8638b724c39824d0"}]}`)
	})
}
