// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

// TODO <>/<> wrap all errors

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	CachedStoreEventID      = "cached_store"
	CachedStorePutMethod    = "put"
	CachedStoreDeleteMethod = "delete"
)

type cachedStoreEvent[T any] struct {
	Key       string    `json:"key"`
	Method    string    `json:"method"`
	SentAt    time.Time `json:"sent_at"`
	StoreName string    `json:"name"`

	Data T `json:"data,omitempty"`
}

type IndexEntry[T any] struct {
	Key       string `json:"k"`
	ValueHash string `json:"h"`
	data      T
}

type StoredIndex[T any] struct {
	Data []IndexEntry[T]
}

type Index[T any] map[string]*IndexEntry[T]

type CachedStore[T any] struct {
	// dependencies
	name   string
	logger utils.Logger
	mmapi  *pluginapi.Client
	papi   plugin.API

	// internal
	cache        *sync.Map // of *IndexEntry[T]
	persistMutex *cluster.Mutex
}

func MakeCachedStore[T any](name string, api plugin.API, mmapi *pluginapi.Client, logger utils.Logger) (*CachedStore[T], error) {
	s := &CachedStore[T]{
		name:   name,
		papi:   api,
		mmapi:  mmapi,
		logger: logger,

		cache: &sync.Map{},
	}

	mutex, err := cluster.NewMutex(api, s.mutexKey())
	if err != nil {
		return nil, err
	}
	s.persistMutex = mutex

	_, err = s.syncFromKV(false)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *CachedStore[T]) Index() Index[T] {
	out := Index[T]{}
	s.cache.Range(func(key, mapv interface{}) bool {
		entry := mapv.(*IndexEntry[T])
		out[entry.Key] = entry
		return true
	})
	return out
}

func (s *CachedStore[T]) get(key string) (entry *IndexEntry[T], ok bool) {
	mapv, ok := s.cache.Load(key)
	if !ok {
		return entry, false
	}
	return mapv.(*IndexEntry[T]), true
}

func (s *CachedStore[T]) Get(key string) (result T, ok bool) {
	mapv, ok := s.cache.Load(key)
	if !ok {
		return result, false
	}
	entry := mapv.(*IndexEntry[T])
	return entry.data, true
}

func (s *CachedStore[T]) withRollbacks(r *incoming.Request, updatef func(prevIndex StoredIndex[T]) (rollbacks []func(), err error)) (err error) {
	s.persistMutex.Lock()
	defer s.persistMutex.Unlock()

	prevIndex, err := s.syncFromKV(true)
	if err != nil {
		return err
	}

	rollbacks, err := updatef(prevIndex)
	if err != nil {
		for _, rollback := range rollbacks {
			rollback()
		}
	}
	return err
}

func (s *CachedStore[T]) Put(r *incoming.Request, key string, value T) (err error) {
	return s.withRollbacks(r, func(prevIndex StoredIndex[T]) (rollbacks []func(), err error) {
		hash := jsonHash(value)
		prevEntry, ok := s.get(key)
		if ok && prevEntry.ValueHash == hash {
			return rollbacks, nil
		}

		err = s.persistItem(key, value)
		if err != nil {
			return rollbacks, err
		}
		rollbacks = append(rollbacks, func() {
			if prevEntry != nil {
				_ = s.persistItem(key, prevEntry.data)
			} else {
				_ = s.deleteItem(key)
			}
		})

		newIndex := s.Index()
		newEntry := IndexEntry[T]{
			Key:       key,
			ValueHash: hash,
			data:      value,
		}
		newIndex[key] = &newEntry

		if err = s.persistIndex(newIndex.Stored()); err != nil {
			r.Log.WithError(err).Warnf("CachedStore.Put: failed to persist index, rolling back to previous state")
			return rollbacks, err
		}
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prevIndex) })

		if err = s.notifyPut(key, value); err != nil {
			r.Log.WithError(err).Warnf("CachedStore.Put: failed to send cluster message, rolling back to previous state")
			return rollbacks, err
		}

		s.cache.Store(key, &newEntry)
		return nil, nil
	})
}

func (s *CachedStore[T]) Delete(r *incoming.Request, key string) (err error) {
	return s.withRollbacks(r, func(prevIndex StoredIndex[T]) (rollbacks []func(), err error) {
		prevEntry, ok := s.get(key)
		if !ok {
			return rollbacks, utils.ErrNotFound
		}

		newIndex := s.Index()
		delete(newIndex, key)
		if err = s.persistIndex(newIndex.Stored()); err != nil {
			return rollbacks, err
		}
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prevIndex) })

		err = s.deleteItem(key)
		if err != nil {
			r.Log.WithError(err).Warnf("CachedStore.Delete: failed to remove the item from KV, rolling back to previous state")
			return rollbacks, err
		}
		if prevEntry != nil {
			rollbacks = append(rollbacks, func() { _ = s.persistItem(key, prevEntry.data) })
		}

		if err = s.notifyDelete(key); err != nil {
			r.Log.WithError(err).Warnf("CachedStore.Delete: failed to send cluster message, rolling back to previous state")
			return rollbacks, err
		}

		s.cache.Delete(key)
		return nil, nil
	})
}

func (s *CachedStore[T]) notifyPut(key string, data T) error {
	event := s.newEvent(CachedStorePutMethod, key, data)
	bb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: CachedStoreEventID, Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *CachedStore[T]) notifyDelete(key string) error {
	var data T
	bb, err := json.Marshal(s.newEvent(CachedStoreDeleteMethod, key, data))
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: CachedStoreEventID, Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *CachedStore[T]) processClusterEvent(event cachedStoreEvent[T]) error {
	s.logger.Debugf("CachedStore.processClusterEvent %s: %s key %s", s.name, event.Method, event.Key)

	switch event.Method {
	case CachedStorePutMethod:
		s.cache.Store(event.Key, &IndexEntry[T]{
			Key:       event.Key,
			ValueHash: jsonHash(event.Data),
			data:      event.Data,
		})

	case CachedStoreDeleteMethod:
		s.cache.Delete(event.Key)
	}

	return nil
}

func (s *CachedStore[T]) newEvent(method, key string, data T) cachedStoreEvent[T] {
	return cachedStoreEvent[T]{
		Method:    method,
		StoreName: s.name,
		Key:       key,
		SentAt:    time.Now(),
		Data:      data,
	}
}

func (s *CachedStore[T]) syncFromKV(logWarnings bool) (prevPersistedIndex StoredIndex[T], err error) {
	var nilIndex StoredIndex[T]
	index := StoredIndex[T]{}
	err = s.mmapi.KV.Get(s.indexKey(), &index)
	if err != nil {
		return nilIndex, err
	}
	prevPersistedIndex = index

	cachedIndex := s.Index().Stored()
	if cachedIndex.hash() != prevPersistedIndex.hash() {
		add, change, remove := cachedIndex.compareTo(prevPersistedIndex)
		if logWarnings {
			s.logger.Warnf("stale cache for %s, rebuilding. extra keys: %v, missing keys:%v, different values for keys:%v", s.name, remove, add, change)
		}

		for _, entry := range remove {
			s.logger.Debugf("CachedStore.synchFromKV %s: key %s in cache but no longer in index, deleting", s.name, entry.Key)
			s.cache.Delete(entry.Key)
		}
		for _, entry := range append(add, change...) {
			entry.data, err = s.getItem(entry.Key)
			if err != nil {
				return nilIndex, err
			}
			storableClone := entry
			s.logger.Debugf("CachedStore.synchFromKV %s: loaded missing or stale key %s", s.name, entry.Key)
			s.cache.Store(entry.Key, &storableClone)
		}
	}

	return prevPersistedIndex, nil
}

func (s *CachedStore[T]) indexKey() string {
	return fmt.Sprintf("%s.%s-index", KVCachedPrefix, s.name)
}

func (s *CachedStore[T]) mutexKey() string {
	return fmt.Sprintf("%s.%s-mutex", KVCachedPrefix, s.name)
}

func (s *CachedStore[T]) itemKey(key string) string {
	return fmt.Sprintf("%s.%s-item-%s", KVCachedPrefix, s.name, key)
}

func (s *CachedStore[T]) persistIndex(index StoredIndex[T]) error {
	_, err := s.mmapi.KV.Set(s.indexKey(), index)
	return err
}

func (s *CachedStore[T]) persistItem(key string, value T) error {
	_, err := s.mmapi.KV.Set(s.itemKey(key), value)
	return err
}

func (s *CachedStore[T]) deleteItem(key string) error {
	return s.mmapi.KV.Delete(s.itemKey(key))
}

func (s *CachedStore[T]) getItem(key string) (T, error) {
	var v T
	err := s.mmapi.KV.Get(s.itemKey(key), &v)
	return v, err
}

func (entry *IndexEntry[T]) String() string {
	return entry.Key
}

func (index Index[T]) Stored() StoredIndex[T] {
	stored := StoredIndex[T]{}
	for _, v := range index {
		stored.Data = append(stored.Data, *v)
	}
	stored.sort()
	return stored
}

func (index StoredIndex[T]) sort() {
	sort.Slice(index.Data, func(i, j int) bool {
		return index.Data[i].Key < index.Data[j].Key
	})
}

func (index StoredIndex[T]) hash() string {
	var b []byte
	for _, item := range index.Data {
		b = append(b, item.Key...)
		b = append(b, item.ValueHash...)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

func (index *StoredIndex[T]) compareTo(other StoredIndex[T]) (add, change, remove []IndexEntry[T]) {
	otherData := other.Data
	if otherData == nil {
		otherData = []IndexEntry[T]{}
	}
	indexData := index.Data
	if indexData == nil {
		indexData = []IndexEntry[T]{}
	}

	i, o := 0, 0
	for {
		switch {
		case i >= len(indexData):
			return append(add, otherData[o:]...), change, remove
		case o >= len(otherData):
			return add, change, append(remove, indexData[i:]...)
		case indexData[i].Key < otherData[o].Key:
			remove = append(remove, indexData[i])
			i++
		case indexData[i].Key > otherData[o].Key:
			add = append(add, otherData[o])
			o++
		default:
			if indexData[i].ValueHash != otherData[o].ValueHash {
				change = append(change, otherData[o])
			}
			i++
			o++
		}
	}
}

func (index Index[T]) Clone() Index[T] {
	out := Index[T]{}
	for k, v := range index {
		out[k] = v
	}
	return out
}

func parseCachedStoreKey(key string) (name, id string, err error) {
	parts := strings.SplitN(key, "-", 3)
	if len(parts) != 3 {
		return "", "", errors.Wrap(utils.ErrInvalid, "cached store item key: "+key)
	}

	id = parts[2]
	parts = strings.Split(parts[0], ".")
	if len(parts) != 3 || parts[0] != "" || "."+parts[1] != KVCachedPrefix {
		return "", "", errors.Wrap(utils.ErrInvalid, "cached store item key: "+key)
	}
	name = parts[2]

	return name, id, nil
}

func jsonHash(value any) string {
	data, _ := json.Marshal(value)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
