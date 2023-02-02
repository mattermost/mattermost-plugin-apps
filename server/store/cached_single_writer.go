// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SingleWriterCachedStore[T Cloneable[T]] struct {
	*SimpleCachedStore[T]
	papi         plugin.API
	iAmTheLeader bool
}

func SingleWriterCachedStoreMaker[T Cloneable[T]](api plugin.API, mmapi *pluginapi.Client, log utils.Logger) func(string) (CachedStore[T], error) {
	return func(name string) (CachedStore[T], error) {
		return MakeSingleWriterCachedStore[T](name, api, mmapi, log)
	}
}

func MakeSingleWriterCachedStore[T Cloneable[T]](name string, api plugin.API, mmapi *pluginapi.Client, log utils.Logger) (*SingleWriterCachedStore[T], error) {
	base, err := MakeSimpleCachedStore[T](name, mmapi, log)
	if err != nil {
		return nil, err
	}
	s := &SingleWriterCachedStore[T]{
		SimpleCachedStore: base,
		papi:              api,
	}

	cachedStoreEventSink.Store(s.eventID(), s.onEvent)
	return s, nil
}

func (s *SingleWriterCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	// tell everyone else about the change
	err := s.notify(s.eventID(), s.newPutEvent(key, value))
	if err != nil {
		r.Log.WithError(err).Warnf("SingleWriterCachedStore: failed to send cluster message, rolling back to previous state")
		return errors.Wrapf(err, "failed to send put cluster message for key %s", key)
	}

	return s.SimpleCachedStore.update(r, true, key, value,
		func(value *T, changed *StoredIndex[T]) error {
			if changed != nil {
				// sync everyone else to the new index hash
				if err := s.notify(s.eventID(), s.newSyncEvent(changed.hash())); err != nil {
					r.Log.WithError(err).Warnf("SingleWriterCachedStore: failed to send cluster message, rolling back to previous state")
					return errors.Wrapf(err, "failed to send sync cluster message")
				}
			}
			return nil
		})
}

func (s *SingleWriterCachedStore[T]) eventID() string {
	return CachedStoreEventID + "/" + s.name
}

func (s *SingleWriterCachedStore[T]) notify(id string, event CachedStoreClusterEvent[T]) error {
	bb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: id, Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *SingleWriterCachedStore[T]) onEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	event := CachedStoreClusterEvent[T]{}
	err := json.Unmarshal(ev.Data, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}
	r.Log.Debugf("SingleWriterCachedStore.processClusterEvent %s: received key %s; new index hash: `%s`", s.name, event.Key, event.IndexHash)

	if event.Key != "" {
		// A put event.
		persist := true
		notifyOthers := func(newValue *T, newStoredIndex *StoredIndex[T]) error {
			return s.notify(s.eventID(), s.newSyncEvent(newStoredIndex.hash()))
		}
		if !s.iAmTheLeader {
			persist = false
			notifyOthers = nil
		}

		if err := s.SimpleCachedStore.update(r, persist, event.Key, event.Value, notifyOthers); err != nil {
			return err
		}
	}

	if event.IndexHash != "" && event.IndexHash != s.Index().Stored().hash() {
		r.Log.Debugf("SingleWriterCachedStore.processClusterEvent %s: index hash mismatch, syncing from KV", s.name)
		if _, err := s.syncFromKV(r.Log, true); err != nil {
			return err
		}
	}

	return nil
}

func (s *SingleWriterCachedStore[T]) newPutEvent(key string, data *T) CachedStoreClusterEvent[T] {
	return CachedStoreClusterEvent[T]{
		Value:     data,
		Key:       key,
		StoreName: s.name,
	}
}

func (s *SingleWriterCachedStore[T]) newSyncEvent(indexHash string) CachedStoreClusterEvent[T] {
	return CachedStoreClusterEvent[T]{
		StoreName: s.name,
	}
}
