// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

type onPluginClusterMessage func(r *incoming.Request, ev model.PluginClusterEvent) error

var cachedStorePluginClusterMessageSink = map[string]onPluginClusterMessage{}

// CachedStoreClusterEvent is a cluster event sent between nodes. It works for
// both the mutex-based and the single writer implementations. If Key is set,
// the message instructs to modify the key. If Data is nil, it is a delete
// operation. If IndexHash set, the receiver attempts to re-sync to it from the
// KV store. differentiate between put and delete (nil) events.
type CachedStoreClusterEvent[T any] struct {
	Value     *T     `json:"value,omitempty"`
	IndexHash string `json:"index_hash,omitempty"`
	Key       string `json:"key"`
	StoreName string `json:"name"`
}

const (
	CachedStoreEventID = "cached_store"
)

func OnPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) {
	f, ok := cachedStorePluginClusterMessageSink[ev.Id]
	if !ok {
		r.Log.Debugf("OnPluginClusterEvent: no handler for %s", ev.Id)
		return
	}
	err := f(r, ev)
	if err != nil {
		r.Log.WithError(err).Errorw("failed to handle plugin cluster event")
	}
}
