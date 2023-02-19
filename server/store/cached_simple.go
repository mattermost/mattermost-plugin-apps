// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type SimpleCachedStore[T Cloneable[T]] struct {
	cache *sync.Map // of *IndexEntry[T]
	api   config.API
	name  string
}

var _ CachedStore[testDataType] = (*SimpleCachedStore[testDataType])(nil)

func MakeSimpleCachedStore[T Cloneable[T]](name string, api config.API, log utils.Logger) (*SimpleCachedStore[T], error) {
	s := &SimpleCachedStore[T]{
		cache: &sync.Map{},
		name:  name,
		api:   api,
	}

	index, _, _, err := s.syncFromKVDetailed()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize a cached store %s from KV", s.name)
	}
	log.Debugf("initialized cached store: %s, %v items", s.name, len(index.Data))
	return s, nil
}

func (s *SimpleCachedStore[T]) Stop() {}

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
	_, _, err := s.update(r, true, key, value)
	return err
}

func (s *SimpleCachedStore[T]) update(r *incoming.Request, persist bool, key string, value *T) (updated *StoredIndex[T], changed bool, err error) {
	var rollbacks []func()
	defer func() {
		if err != nil {
			for i := len(rollbacks) - 1; i >= 0; i-- {
				rollbacks[i]()
			}
			err = errors.Wrapf(err, "failed to update cached store %s, rolled back", key)
		}
	}()

	prev := s.Index().Stored()
	prevEntry := s.getEntry(key)
	var newEntry *CachedIndexEntry[T]
	if value == nil {
		if prevEntry == nil {
			// No change.
			return prev, false, nil
		}

		if persist {
			err = s.deleteItem(key)
			if err != nil {
				return nil, false, errors.Wrap(err, "failed to delete item")
			}
			rollbacks = append(rollbacks, func() {
				_ = s.persistItem(key, prevEntry.data)
			})
		}
	} else {
		valueHash := valueHash(value)
		if prevEntry != nil && prevEntry.ValueHash == valueHash {
			// No change.
			return prev, false, nil
		}

		if persist {
			err = s.persistItem(key, value)
			if err != nil {
				return nil, false, errors.Wrap(err, "failed to persist item")
			}
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
			data:      value,
		}
	}

	s.updateCachedValue(key, newEntry)
	rollbacks = append(rollbacks, func() { s.updateCachedValue(key, prevEntry) })

	stored := s.Index().Stored()
	if persist {
		if err = s.persistIndex(stored); err != nil {
			return nil, false, errors.Wrap(err, "failed to persist index")
		}
		rollbacks = append(rollbacks, func() { _ = s.persistIndex(prev) })
	}

	return stored, true, nil
}

func (s *SimpleCachedStore[T]) updateCachedValue(key string, value *CachedIndexEntry[T]) {
	if value != nil {
		s.cache.Store(key, value)
	} else {
		s.cache.Delete(key)
	}
}

func (s *SimpleCachedStore[T]) syncFromKV() error {
	_, _, _, err := s.syncFromKVDetailed()
	return err
}

func (s *SimpleCachedStore[T]) syncFromKVDetailed() (stored *StoredIndex[T], changed, removed []CachedIndexEntry[T], err error) {
	stored = &StoredIndex[T]{}
	err = s.api.Mattermost.KV.Get(s.indexKey(), stored)
	if err != nil {
		return nil, nil, nil, err
	}

	// compare the current cache with the index, update to match
	changed, removed = s.Index().Stored().compareTo(stored)
	if len(changed) == 0 && len(removed) == 0 {
		return stored, changed, removed, nil
	}

	for _, entry := range removed {
		s.cache.Delete(entry.Key)
	}
	for _, entry := range changed {
		entry.data, err = s.getItem(entry.Key)
		if err != nil {
			return nil, nil, nil, err
		}

		storableClone := entry
		s.cache.Store(entry.Key, &storableClone)
	}

	return stored, changed, removed, nil
}

func (s *SimpleCachedStore[T]) indexKey() string {
	return fmt.Sprintf("%s.%s-index", CachedPrefix, s.name)
}

func (s *SimpleCachedStore[T]) itemKey(key string) string {
	return fmt.Sprintf("%s.%s-item-%s", CachedPrefix, s.name, key)
}

func (s *SimpleCachedStore[T]) persistIndex(index *StoredIndex[T]) error {
	_, err := s.api.Mattermost.KV.Set(s.indexKey(), index)
	return err
}

func (s *SimpleCachedStore[T]) persistItem(key string, value *T) error {
	_, err := s.api.Mattermost.KV.Set(s.itemKey(key), value)
	return err
}

func (s *SimpleCachedStore[T]) deleteItem(key string) error {
	return s.api.Mattermost.KV.Delete(s.itemKey(key))
}

func (s *SimpleCachedStore[T]) getItem(key string) (*T, error) {
	var v T
	err := s.api.Mattermost.KV.Get(s.itemKey(key), &v)
	return &v, err
}

func parseCachedStoreKey(key string) (name, id string, err error) {
	parts := strings.SplitN(key, "-", 3)
	if len(parts) != 3 {
		return "", "", errors.Wrap(utils.ErrInvalid, "cached store item key: "+key)
	}

	id = parts[2]
	parts = strings.Split(parts[0], ".")
	if len(parts) != 3 || parts[0] != "" || "."+parts[1] != CachedPrefix {
		return "", "", errors.Wrap(utils.ErrInvalid, "cached store item key: "+key)
	}
	name = parts[2]

	return name, id, nil
}

func valueHash(value any) string {
	data, _ := json.Marshal(value)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
