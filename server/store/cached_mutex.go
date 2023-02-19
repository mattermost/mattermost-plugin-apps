// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// MutexCachedStore is a CachedStore that uses a cluster.Mutex to ensure only
// one of the nodes is writing to the KV store at a time. Other nodes get
// notifies with [CachedStoreClusterEvent] messages and update their in-memory
// caches accordingly.
type MutexCachedStore[T Cloneable[T]] struct {
	*SimpleCachedStore[T]
	cluster *CachedStoreCluster
	kvMutex *cluster.Mutex
}

func MakeMutexCachedStore[T Cloneable[T]](name string, c *CachedStoreCluster, log utils.Logger) (CachedStore[T], error) {
	return makeMutexCachedStore[T](name, c, log)
}

func makeMutexCachedStore[T Cloneable[T]](name string, c *CachedStoreCluster, log utils.Logger) (*MutexCachedStore[T], error) {
	base, err := MakeSimpleCachedStore[T](name, c.api, log)
	if err != nil {
		return nil, err
	}
	s := &MutexCachedStore[T]{
		SimpleCachedStore: base,
		cluster:           c,
	}

	mutex, err := cluster.NewMutex(s.api.Plugin, s.mutexKey())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make a new cached store %s", s.name)
	}
	s.kvMutex = mutex

	c.setEventHandler(s.eventID(), s.onEvent)

	return s, nil
}

func (s *MutexCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	// Ensure exclusive access to the KV store.
	s.kvMutex.Lock()

	// Need to ensure the local index is up to date before we can update it.
	if err := s.syncFromKV(); err != nil {
		s.kvMutex.Unlock()
		return errors.Wrapf(err, "failed to sync from KV store before updating key %s", key)
	}

	updatedIndex, changed, err := s.SimpleCachedStore.update(r, true, key, value)
	s.kvMutex.Unlock()
	if err != nil {
		return err
	}

	r.Log.Debugf("<>/<> 1 %v", changed)

	if changed {
		event := s.newPluginClusterEvent(key, value, updatedIndex.hash())
		s.cluster.broadcastEvent(r, s.eventID(), event)
	}
	return nil
}

func (s *MutexCachedStore[T]) eventID() string {
	return putEventID + "/" + s.name
}

func (s *MutexCachedStore[T]) onEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	event := cachedStoreClusterEvent[T]{}
	err := json.Unmarshal(ev.Data, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}

	var newEntry *CachedIndexEntry[T]
	if event.Value != nil {
		newEntry = &CachedIndexEntry[T]{
			Key:       event.Key,
			ValueHash: valueHash(event.Value),
			data:      event.Value,
		}
	}
	s.SimpleCachedStore.updateCachedValue(event.Key, newEntry)

	h := s.Index().Stored().hash()
	if h != event.IndexHash {
		r.Log.Debugf("cluster event %s: %s: updated index hash `%s` mismatched expected `%s`, syncing from KV", s.name, ev.Id, h, event.IndexHash)
		index, _, _, err := s.syncFromKVDetailed()
		if err != nil {
			return err
		}
		if h = index.hash(); h != event.IndexHash {
			r.Log.Debugf("cluster event %s: %s: synced from KV, still hash mismatch: got `%s`, expected `%s`", s.name, ev.Id, h, event.IndexHash)
		}
	} else {
		r.Log.Debugf("cluster event %s: %s: key has %s been updated", s.name, ev.Id, event.Key)
	}
	return nil
}

func (s *MutexCachedStore[T]) newPluginClusterEvent(key string, data *T, indexHash string) cachedStoreClusterEvent[T] {
	return cachedStoreClusterEvent[T]{
		StoreName: s.name,
		Key:       key,
		Value:     data,
		IndexHash: indexHash,
	}
}

func (s *MutexCachedStore[T]) mutexKey() string {
	return fmt.Sprintf("%s.%s-mutex", CachedPrefix, s.name)
}
