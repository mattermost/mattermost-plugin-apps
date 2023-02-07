// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"fmt"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type MutexCachedStore[T Cloneable[T]] struct {
	*SimpleCachedStore[T]
	papi    plugin.API
	kvMutex *cluster.Mutex
}

func MutexCachedStoreMaker[T Cloneable[T]](api plugin.API, mmapi *pluginapi.Client, log utils.Logger) func(string) (CachedStore[T], error) {
	return func(name string) (CachedStore[T], error) {
		return MakeMutexCachedStore[T](name, api, mmapi, log)
	}
}

func MakeMutexCachedStore[T Cloneable[T]](name string, api plugin.API, mmapi *pluginapi.Client, log utils.Logger) (*MutexCachedStore[T], error) {
	base, err := MakeSimpleCachedStore[T](name, mmapi, log)
	if err != nil {
		return nil, err
	}
	s := &MutexCachedStore[T]{
		SimpleCachedStore: base,
		papi:              api,
	}

	mutex, err := cluster.NewMutex(api, s.mutexKey())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make a new cached store %s", s.name)
	}
	s.kvMutex = mutex

	cachedStorePluginClusterMessageSink[s.eventID()] = s.onEvent
	return s, nil
}

func (s *MutexCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	s.kvMutex.Lock()
	defer s.kvMutex.Unlock()

	return s.SimpleCachedStore.update(r, true, key, value,
		func(value *T, changed *StoredIndex[T]) error {
			if changed != nil {
				if err := s.notify(key, value, changed.hash()); err != nil {
					r.Log.WithError(err).Warnf("MutexCachedStore: failed to send cluster message, rolling back to previous state")
					return errors.Wrapf(err, "failed to send cluster message for key %s", key)
				}
			}
			return nil
		})
}

func (s *MutexCachedStore[T]) eventID() string {
	return CachedStoreEventID + "/" + s.name
}

func (s *MutexCachedStore[T]) notify(key string, data *T, indexHash string) error {
	event := s.newPluginClusterEvent(key, data, indexHash)
	bb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: s.eventID(), Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *MutexCachedStore[T]) onEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	event := CachedStoreClusterEvent[T]{}
	err := json.Unmarshal(ev.Data, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}
	r.Log.Debugf("MutexCachedStore.processClusterEvent %s: received for %s; new index hash: `%s`", s.name, event.Key, event.IndexHash)

	var newEntry *CachedIndexEntry[T]
	if event.Value != nil {
		newEntry = &CachedIndexEntry[T]{
			Key:       event.Key,
			ValueHash: jsonHash(event.Value),
			data:      event.Value,
		}
	}
	s.SimpleCachedStore.updateCachedValue(event.Key, newEntry)

	if s.Index().Stored().hash() != event.IndexHash {
		r.Log.Debugf("MutexCachedStore.processClusterEvent %s: index hash mismatch, syncing from KV", s.name)
		if _, err := s.syncFromKV(r.Log, true); err != nil {
			return err
		}
	} else {
		r.Log.Debugf("MutexCachedStore.processClusterEvent %s: %s key %s", s.name, event.Key)
	}
	return nil
}

func (s *MutexCachedStore[T]) newPluginClusterEvent(key string, data *T, indexHash string) CachedStoreClusterEvent[T] {
	return CachedStoreClusterEvent[T]{
		StoreName: s.name,
		Key:       key,
		Value:     data,
		IndexHash: indexHash,
	}
}

func (s *MutexCachedStore[T]) mutexKey() string {
	return fmt.Sprintf("%s.%s-mutex", KVCachedPrefix, s.name)
}
