// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Cloneable[T any] interface {
	Clone() *T
}

// CachedStore is a cluster-aware write-through in-memory cache for entities such as [Apps] or [Subscriptions]
type CachedStore[T Cloneable[T]] interface {
	Index() CachedIndex[T]
	Get(key string) (value *T)
	Put(r *incoming.Request, key string, value *T) error
	Stop()
}

func MakeCachedStore[T Cloneable[T]](name string, cluster *CachedStoreCluster, log utils.Logger) (CachedStore[T], error) {
	return makeCachedStore[T](cluster.defaultStoreKind, name, cluster, log)
}

func makeCachedStore[T Cloneable[T]](kind CachedStoreClusterKind, name string, cluster *CachedStoreCluster, log utils.Logger) (CachedStore[T], error) {
	switch kind {
	case MutexCachedStoreKind:
		return MakeMutexCachedStore[T](name, cluster, log)
	case SingleWriterCachedStoreKind:
		return MakeSingleWriterCachedStore[T](name, cluster, log)
	case SimpleCachedStoreKind:
		return MakeSimpleCachedStore[T](name, cluster.api, log)
	case TestCachedStoreKind:
		return TestingCachedStore[T]{}, nil
	default:
		return nil, fmt.Errorf("unknown CachedStore kind %s", kind)
	}
}

type CachedIndexEntry[T any] struct {
	Key       string `json:"k"`
	ValueHash string `json:"h"`
	data      *T
}

type StoredIndex[T any] struct {
	Data []CachedIndexEntry[T]
}

type CachedIndex[T any] map[string]*CachedIndexEntry[T]

func (index *StoredIndex[T]) String() string {
	keys := []string{}
	for _, item := range index.Data {
		keys = append(keys, item.Key)
	}
	return fmt.Sprintf("%v `%s` %v", len(index.Data), utils.FirstN(index.hash(), 10), keys)
}

func (entry *CachedIndexEntry[T]) String() string {
	return entry.Key + "/" + entry.ValueHash
}

func (index CachedIndex[T]) Stored() *StoredIndex[T] {
	stored := &StoredIndex[T]{}
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

func (index *StoredIndex[T]) compareTo(other *StoredIndex[T]) (change, remove []CachedIndexEntry[T]) {
	otherData := other.Data
	if otherData == nil {
		otherData = []CachedIndexEntry[T]{}
	}
	indexData := index.Data
	if indexData == nil {
		indexData = []CachedIndexEntry[T]{}
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
