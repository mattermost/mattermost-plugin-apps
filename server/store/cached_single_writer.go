// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// SingleWriterCachedStore is a CachedStore that ensures only the cluster leader
// node is writing to the KV store. If the request to modify the store
// originates on a non-leader node, it broadcasts a [CachedStoreClusterEvent] to
// all other nodes, including the leader which persists the data. The leader
// then broadcasts the updated index SHA to all other nodes that can re-load
// from KV if out of sync.
type SingleWriterCachedStore[T Cloneable[T]] struct {
	*SimpleCachedStore[T]
	cluster *CachedStoreCluster

	notifyTimerMutex *sync.Mutex
	notifyTimer      *time.Timer
}

func MakeSingleWriterCachedStore[T Cloneable[T]](name string, c *CachedStoreCluster, log utils.Logger) (CachedStore[T], error) {
	base, err := MakeSimpleCachedStore[T](name, c.api, log)
	if err != nil {
		return nil, err
	}
	s := &SingleWriterCachedStore[T]{
		SimpleCachedStore: base,
		cluster:           c,
		notifyTimerMutex:  &sync.Mutex{},
	}

	c.setEventHandler(s.getPutEventID(), s.onPutEvent)
	c.setEventHandler(s.getSyncEventID(), s.onSyncEvent)

	return s, nil
}

const syncBroadcastDelay = 500 * time.Millisecond

func (s *SingleWriterCachedStore[T]) broadcastSyncLater(r *incoming.Request, hash string) {
	s.notifyTimerMutex.Lock()
	defer s.notifyTimerMutex.Unlock()

	if s.notifyTimer != nil {
		s.notifyTimer.Stop()
	}
	r = r.Clone()
	s.notifyTimer = time.AfterFunc(syncBroadcastDelay, func() {
		s.cluster.broadcastEvent(r, s.getSyncEventID(), s.newSyncEvent(hash))
	})
}

func (s *SingleWriterCachedStore[T]) notifyOthersOnPut(r *incoming.Request, isOrigin, isLeader bool, key string, value *T, newIndexSHA string) {
	// The host that originated the change broadcasts the new data to
	// everyone else.
	if isOrigin {
		s.cluster.broadcastEvent(r, s.getPutEventID(), s.newPutEvent(key, value, newIndexSHA))
	}

	// The host that persisted the data broadcasts the new index hash
	// for others to ensure consistency.
	if isLeader {
		s.broadcastSyncLater(r, newIndexSHA)
	}
}

func (s *SingleWriterCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	isLeader := r.Config.IsClusterLeader()
	isOrigin := true

	index, changed, err := s.SimpleCachedStore.update(r, isLeader, key, value)
	if !changed {
		return err
	}

	s.notifyOthersOnPut(r, isOrigin, isLeader, key, value, index.hash())
	return nil
}

func (s *SingleWriterCachedStore[T]) getPutEventID() string {
	return putEventID + "/" + s.name
}

func (s *SingleWriterCachedStore[T]) getSyncEventID() string {
	return syncEventID + "/" + s.name
}

func (s *SingleWriterCachedStore[T]) onPutEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	event := cachedStoreClusterEvent[T]{}
	err := json.Unmarshal(ev.Data, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}
	if event.Key == "" {
		return utils.NewInvalidError("key is empty")
	}
	isClusterLeader := r.Config.IsClusterLeader()
	// r.Log.Debugf("received cluster event: PUT %s: %s; new index hash: `%s`, isClusterLeader: %t",
	// s.name, event.Key, utils.FirstN(event.IndexHash, 10), isClusterLeader)

	if !isClusterLeader {
		// if not the leader, update the in-memory local cache only.
		_, _, err = s.SimpleCachedStore.update(r, false, event.Key, event.Value)
		return err
	}

	index, changed, err := s.SimpleCachedStore.update(r, true, event.Key, event.Value)
	if !changed {
		return err
	}

	s.notifyOthersOnPut(r, false, isClusterLeader, event.Key, event.Value, index.hash())
	return nil
}

func (s *SingleWriterCachedStore[T]) onSyncEvent(r *incoming.Request, ev model.PluginClusterEvent) error {
	event := cachedStoreClusterEvent[T]{}
	err := json.Unmarshal(ev.Data, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}
	isClusterLeader := r.Config.IsClusterLeader()
	r.Log.Debugf("received cluster event SYNC %s: %s; new index hash: `%s`, isClusterLeader: %t",
		s.name, event.Key, utils.FirstN(event.IndexHash, 10), isClusterLeader)
	if isClusterLeader {
		return errors.New("cluster leader should not receive sync events")
	}
	if event.IndexHash == "" {
		return utils.NewInvalidError("IndexHash is empty")
	}

	if event.IndexHash != s.Index().Stored().hash() {
		r.Log.Debugf("SingleWriterCachedStore.processClusterEvent %s: index hash mismatch, syncing from KV", s.name)
		updated, _, _, err := s.syncFromKVDetailed()
		if err != nil {
			return err
		}
		if event.IndexHash != updated.hash() {
			return errors.Errorf("index hash mismatch after sync: `%s`, expected `%s`", event.IndexHash, updated.hash())
		}
	}
	return nil
}

func (s *SingleWriterCachedStore[T]) newPutEvent(key string, data *T, indexHash string) cachedStoreClusterEvent[T] {
	return cachedStoreClusterEvent[T]{
		Value:     data,
		Key:       key,
		StoreName: s.name,
		IndexHash: indexHash,
	}
}

func (s *SingleWriterCachedStore[T]) newSyncEvent(indexHash string) cachedStoreClusterEvent[T] {
	return cachedStoreClusterEvent[T]{
		StoreName: s.name,
		IndexHash: indexHash,
	}
}
