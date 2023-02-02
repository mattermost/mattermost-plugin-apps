// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SimpleCachedStore[T Cloneable[T]] struct {
	// dependencies
	mmapi *pluginapi.Client

	// internal
	name  string
	cache *sync.Map // of *IndexEntry[T]
}

var _ CachedStore[apps.App] = (*SimpleCachedStore[apps.App])(nil)

func SimpleCachedStoreMaker[T Cloneable[T]](mmapi *pluginapi.Client, log utils.Logger) func(string) (CachedStore[T], error) {
	return func(name string) (CachedStore[T], error) {
		return MakeSimpleCachedStore[T](name, mmapi, log)
	}
}

func MakeSimpleCachedStore[T Cloneable[T]](name string, mmapi *pluginapi.Client, log utils.Logger) (*SimpleCachedStore[T], error) {
	s := &SimpleCachedStore[T]{
		name:  name,
		mmapi: mmapi,
		cache: &sync.Map{},
	}

	log.Debugf("initializing cached store: %s", s.name)
	_, err := s.syncFromKV(log, false)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize a cached store %s from KV", s.name)
	}
	return s, nil
}

func (s *SimpleCachedStore[T]) Index() CachedIndex[T] {
	out := CachedIndex[T]{}
	s.cache.Range(func(key, mapv interface{}) bool {
		entry := mapv.(*CachedIndexEntry[T])
		out[entry.Key] = entry
		return true
	})
	return out
}

func (s *SimpleCachedStore[T]) getEntry(key string) *CachedIndexEntry[T] {
	mapv, _ := s.cache.Load(key)
	if mapv == nil {
		return nil
	}
	return mapv.(*CachedIndexEntry[T])
}

func (s *SimpleCachedStore[T]) Get(key string) *T {
	entry := s.getEntry(key)
	if entry == nil {
		return nil
	}
	return entry.data
}

// update returns nil for newIndex if there was no change
func (s *SimpleCachedStore[T]) Put(r *incoming.Request, key string, value *T) error {
	return s.update(r, true, key, value, nil)
}

func (s *SimpleCachedStore[T]) update(r *incoming.Request, persist bool, key string, value *T, notify func(newValue *T, newStoredIndex *StoredIndex[T]) error) (err error) {
	var rollbacks []func()
	defer func() {
		if err != nil {
			for i := len(rollbacks) - 1; i >= 0; i-- {
				rollbacks[i]()
			}
			err = errors.Wrapf(err, "failed to update cached store %s, rolled back", key)
		}
	}()

	newIndex := s.Index()
	prevStoredIndex := newIndex.Stored()
	if persist {
		prevStoredIndex, err = s.syncFromKV(r.Log, true)
		if err != nil {
			return errors.Wrapf(err, "failed to sync cached store %s before updating", s.name)
		}
	}

	var newEntry *CachedIndexEntry[T]
	op := ""
	prevEntry := s.getEntry(key)
	if value == nil {
		op = fmt.Sprintf("cachedStore.Delete: %s: %s", s.name, key)
		if prevEntry == nil {
			r.Log.Debugf("%s: %s: no change", op, key)
			return nil
		}

		if persist {
			err = s.deleteItem(key)
			if err != nil {
				return errors.Wrap(err, "failed to delete item")
			}
			r.Log.Debugf(" %s: key %s deleted from KV", op, key)
			rollbacks = append(rollbacks, func() {
				_ = s.persistItem(key, prevEntry.data)
			})
		}

		delete(newIndex, key)
	} else {
		op = fmt.Sprintf("cachedStore.Store: %s: %s", s.name, key)
		valueHash := jsonHash(value)
		if prevEntry != nil && prevEntry.ValueHash == valueHash {
			r.Log.Debugf("%s: %s: no change", op, key)
			return nil
		}

		if persist {
			err = s.persistItem(key, value)
			if err != nil {
				return errors.Wrap(err, "failed to persist item")
			}
			r.Log.Debugf(" %s: key %s persisted to KV", op, key)
			rollbacks = append(rollbacks, func() {
				if prevEntry != nil {
					_ = s.persistItem(key, prevEntry.data)
				} else {
					_ = s.deleteItem(key)
				}
			})
		}

		newEntry = &CachedIndexEntry[T]{
			Key:       key,
			ValueHash: valueHash,
			data:      (*value).Clone(),
		}
		newIndex[key] = newEntry
	}

	newStoredIndex := newIndex.Stored()
	if newStoredIndex.hash() == prevStoredIndex.hash() {
		r.Log.Debugf("%s: %s: no change (index)", op, key)
		return nil
	}

	if persist {
		if err = s.persistIndex(newStoredIndex); err != nil {
			r.Log.WithError(err).Warnf("%s: failed to persist index, rolling back to previous state", op)
			return errors.Wrap(err, "failed to persist index")
		}
		r.Log.Debugf("%s: index persisted to KV", op)
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prevStoredIndex) })
	}

	s.updateCachedValue(key, newEntry)
	rollbacks = append(rollbacks, func() { s.updateCachedValue(key, prevEntry) })

	if notify != nil {
		err = notify(value, newStoredIndex)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SimpleCachedStore[T]) updateCachedValue(key string, value *CachedIndexEntry[T]) {
	if value != nil {
		s.cache.Store(key, value)
	} else {
		s.cache.Delete(key)
	}
}

func (s *SimpleCachedStore[T]) syncFromKV(log utils.Logger, logWarnings bool) (prevPersistedIndex *StoredIndex[T], err error) {
	if log == nil {
		log = utils.NilLogger{}
	}
	index := &StoredIndex[T]{}
	err = s.mmapi.KV.Get(s.indexKey(), &index)
	if err != nil {
		return nil, err
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
				return nil, err
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

func (s *SimpleCachedStore[T]) indexKey() string {
	return fmt.Sprintf("%s.%s-index", KVCachedPrefix, s.name)
}

func (s *SimpleCachedStore[T]) itemKey(key string) string {
	return fmt.Sprintf("%s.%s-item-%s", KVCachedPrefix, s.name, key)
}

func (s *SimpleCachedStore[T]) persistIndex(index *StoredIndex[T]) error {
	_, err := s.mmapi.KV.Set(s.indexKey(), index)
	return err
}

func (s *SimpleCachedStore[T]) persistItem(key string, value *T) error {
	_, err := s.mmapi.KV.Set(s.itemKey(key), value)
	return err
}

func (s *SimpleCachedStore[T]) deleteItem(key string) error {
	return s.mmapi.KV.Delete(s.itemKey(key))
}

func (s *SimpleCachedStore[T]) getItem(key string, log utils.Logger) (*T, error) {
	var v T
	err := s.mmapi.KV.Get(s.itemKey(key), &v)
	return &v, err
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
