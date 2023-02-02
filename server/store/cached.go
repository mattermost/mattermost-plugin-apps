// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

// TODO <>/<> wrap all errors

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

type Cloneable[T any] interface {
	Clone() *T
}

type CachedStore[T Cloneable[T]] interface {
	Index() CachedIndex[T]
	Get(key string) (value *T)
	Put(r *incoming.Request, key string, value *T) error
}

type cachedStoreEventProcessor interface {
	OnPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) error
	PluginClusterEventID() string
}

var cachedStoreEventSink = sync.Map{} // of cachedStoreEventProcessor

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
	return fmt.Sprintf("%v", index.Data)
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

func (index CachedIndex[T]) Clone() CachedIndex[T] {
	out := CachedIndex[T]{}
	for k, v := range index {
		out[k] = v
	}
	return out
}

func OnPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) {
	v, ok := cachedStoreEventSink.Load(ev.Id)
	if !ok {
		r.Log.Debugf("OnPluginClusterEvent: no processor event for %s", ev.Id)
		return
	}
	handler, ok := v.(cachedStoreEventProcessor)
	if !ok {
		r.Log.Debugf("OnPluginClusterEvent: invalid event processor %s, type %T", ev.Id, v)
		return
	}

	err := handler.OnPluginClusterEvent(r, ev)
	if err != nil {
		r.Log.WithError(err).Errorw("failed to handle plugin cluster event")
	}
}
