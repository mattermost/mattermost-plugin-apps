// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

// TODO <>/<> wrap all errors

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/cluster"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	CachedStoreEventID      = "cached_store"
	CachedStorePutMethod    = "put"
	CachedStoreDeleteMethod = "delete"
)

type CachedStoreClusterEvent[T any] struct {
	Data      T      `json:"data,omitempty"`
	IndexHash string `json:"index_hash,omitempty"`
	Key       string `json:"key"`
	Method    string `json:"method"`
	StoreName string `json:"name"`
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
	name string
	// logger utils.Logger
	mmapi *pluginapi.Client
	papi  plugin.API

	// internal
	cache   *sync.Map // of *IndexEntry[T]
	kvMutex *cluster.Mutex
}

func MakeCachedStore[T any](name string, api plugin.API, conf config.Service) (*CachedStore[T], error) {
	s := &CachedStore[T]{
		name:  name,
		papi:  api,
		mmapi: conf.MattermostAPI(),
		cache: &sync.Map{},
	}

	mutex, err := cluster.NewMutex(api, s.mutexKey())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make a new cached store %s", s.name)
	}
	s.kvMutex = mutex

	log := s.Log(conf.NewBaseLogger())
	log.Debugf("initializing cached store: %s", s.name)
	_, err = s.syncFromKV(log, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync a new cached store %s from KV", s.name)
	}

	return s, nil
}

func (s *CachedStore[T]) Log(log utils.Logger) utils.Logger {
	hostname, _ := os.Hostname()
	return log.With("cached_store", s.name, "hostname", hostname)
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

func (s *CachedStore[T]) update(log utils.Logger, updatef func(prevIndex StoredIndex[T]) (rollbacks []func(), err error)) (err error) {
	s.kvMutex.Lock()
	defer s.kvMutex.Unlock()

	prevIndex, err := s.syncFromKV(log, true)
	if err != nil {
		return errors.Wrapf(err, "failed to sync cached store %s", s.name)
	}

	rollbacks, err := updatef(prevIndex)
	if err != nil {
		for i := len(rollbacks) - 1; i >= 0; i-- {
			rollbacks[i]()
		}
	}
	return err
}

func (s *CachedStore[T]) Put(r *incoming.Request, key string, value T) (err error) {
	log := s.Log(r.Log)
	return s.update(log, func(prevIndex StoredIndex[T]) (rollbacks []func(), err error) {
		valueHash := jsonHash(value)
		prevEntry, ok := s.get(key)
		if ok && prevEntry.ValueHash == valueHash {
			log.Debugf("CachedStore.Put: %s: %s: no change", s.name, key)
			return rollbacks, nil
		}

		err = s.persistItem(key, value)
		if err != nil {
			return rollbacks, errors.Wrapf(err, "CachedStore.Put: failed to store item %s to %s", key, s.name)
		}
		log.Debugf("CachedStore.Put: %s: %s: persisted to KV", s.name, key)
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
			ValueHash: valueHash,
			data:      value,
		}
		newIndex[key] = &newEntry

		if err = s.persistIndex(newIndex.Stored()); err != nil {
			log.WithError(err).Warnf("CachedStore.Put: failed to persist index, rolling back to previous state")
			return rollbacks, errors.Wrapf(err, "CachedStore.Put: failed to store index to %s", s.name)
		}
		log.Debugf("CachedStore.Put: %s: index persisted to KV", s.name)
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prevIndex) })

		if err = s.notifyPut(key, value); err != nil {
			log.WithError(err).Warnf("CachedStore.Put: failed to send cluster message, rolling back to previous state")
			return rollbacks, errors.Wrapf(err, "CachedStore.Put: failed to send cluster message for key %s in %s", key, s.name)
		}

		s.cache.Store(key, &newEntry)
		return nil, nil
	})
}

func (s *CachedStore[T]) Delete(r *incoming.Request, key string) (err error) {
	log := s.Log(r.Log)
	return s.update(log, func(prevIndex StoredIndex[T]) (rollbacks []func(), err error) {
		prevEntry, ok := s.get(key)
		if !ok {
			return rollbacks, errors.Wrap(utils.ErrNotFound, key)
		}

		newIndex := s.Index()
		delete(newIndex, key)
		if err = s.persistIndex(newIndex.Stored()); err != nil {
			return rollbacks, err
		}
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prevIndex) })

		err = s.deleteItem(key)
		if err != nil {
			log.WithError(err).Warnf("CachedStore.Delete: failed to remove the item from KV, rolling back to previous state")
			return rollbacks, errors.Wrapf(err, "CachedStore.Delete: failed to store index to %s", s.name)
		}
		if prevEntry != nil {
			rollbacks = append(rollbacks, func() { _ = s.persistItem(key, prevEntry.data) })
		}

		if err = s.notifyDelete(key); err != nil {
			log.WithError(err).Warnf("CachedStore.Delete: failed to send cluster message, rolling back to previous state")
			return rollbacks, errors.Wrapf(err, "CachedStore.Delete: failed to send cluster message for key %s in %s", key, s.name)
		}

		s.cache.Delete(key)
		return nil, nil
	})
}

func (s *CachedStore[T]) clusterEventID() string {
	return CachedStoreEventID + "/" + s.name
}

func (s *CachedStore[T]) notifyPut(key string, data T) error {
	event := s.newPluginClusterEvent(CachedStorePutMethod, key, data)
	bb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: s.clusterEventID(), Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *CachedStore[T]) notifyDelete(key string) error {
	var data T
	bb, err := json.Marshal(s.newPluginClusterEvent(CachedStoreDeleteMethod, key, data))
	if err != nil {
		return err
	}

	return s.papi.PublishPluginClusterEvent(
		model.PluginClusterEvent{Id: s.clusterEventID(), Data: bb},
		model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
	)
}

func (s *CachedStore[T]) processClusterEvent(r *incoming.Request, eventData []byte) error {
	event := CachedStoreClusterEvent[T]{}
	err := json.Unmarshal(eventData, &event)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal cached store cluster event")
	}

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

	if s.Index().Stored().hash() != event.IndexHash {
		r.Log.Debugf("CachedStore.processClusterEvent %s: index hash mismatch, syncing from KV", s.name)
		if _, err := s.syncFromKV(r.Log, true); err != nil {
			return err
		}
	} else {
		r.Log.Debugf("CachedStore.processClusterEvent %s: %s key %s", s.name, event.Method, event.Key)
	}
	return nil
}

func (s *CachedStore[T]) newPluginClusterEvent(method, key string, data T) CachedStoreClusterEvent[T] {
	return CachedStoreClusterEvent[T]{
		Method:    method,
		StoreName: s.name,
		Key:       key,
		Data:      data,
		IndexHash: s.Index().Stored().hash(),
	}
}

func (s *CachedStore[T]) syncFromKV(log utils.Logger, logWarnings bool) (prevPersistedIndex StoredIndex[T], err error) {
	if log == nil {
		log = utils.NilLogger{}
	}
	var nilIndex StoredIndex[T]
	index := StoredIndex[T]{}
	err = s.mmapi.KV.Get(s.indexKey(), &index)
	if err != nil {
		return nilIndex, err
	}
	prevPersistedIndex = index

	cachedIndex := s.Index().Stored()
	if cachedIndex.hash() != prevPersistedIndex.hash() {
		change, remove := cachedIndex.compareTo(prevPersistedIndex)
		if logWarnings && (len(change) > 0 || len(remove) > 0) {
			log.Warnf("stale cache for %s, rebuilding. removing keys: %v, updating keys:%v", s.name, remove, change)
		}

		for _, entry := range remove {
			log.Debugf("CachedStore.synchFromKV %s: key %s in cache but no longer in index, deleting", s.name, entry.Key)
			s.cache.Delete(entry.Key)
		}
		for _, entry := range change {
			entry.data, err = s.getItem(entry.Key, log)
			if err != nil {
				return nilIndex, err
			}

			storableClone := entry
			log.Debugf("CachedStore.synchFromKV %s: loaded missing or stale key %s", s.name, entry.Key)
			s.cache.Store(entry.Key, &storableClone)
		}
	} else {
		log.Debugf("CachedStore.synchFromKV %s: cache is up to date", s.name)
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

func (s *CachedStore[T]) getItem(key string, log utils.Logger) (T, error) {
	var v T
	err := s.mmapi.KV.Get(s.itemKey(key), &v)
	log.With("value", v).Debugf("CachedStore.synchFromKV %s: loaded item from KV %s %s", s.name, key, s.itemKey(key))
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

func (index *StoredIndex[T]) compareTo(other StoredIndex[T]) (change, remove []IndexEntry[T]) {
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
			return append(change, otherData[o:]...), remove
		case o >= len(otherData):
			return change, append(remove, indexData[i:]...)
		case indexData[i].Key < otherData[o].Key:
			remove = append(remove, indexData[i])
			i++
		case indexData[i].Key > otherData[o].Key:
			change = append(change, otherData[o])
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
