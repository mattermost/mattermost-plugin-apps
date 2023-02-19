// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package store

import (
	"encoding/json"
	"sync"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type CachedStoreClusterKind string

const (
	MutexCachedStoreKind        = CachedStoreClusterKind("mutex")
	SimpleCachedStoreKind       = CachedStoreClusterKind("simple")
	SingleWriterCachedStoreKind = CachedStoreClusterKind("cluster_leader")
	TestCachedStoreKind         = CachedStoreClusterKind("test")
)

type CachedStoreCluster struct {
	api              config.API
	eventHandlers    *sync.Map // event name -> pluginClusterEventHandler
	defaultStoreKind CachedStoreClusterKind
}

type eventHandler func(r *incoming.Request, ev model.PluginClusterEvent) error

const (
	putEventID  = "cached_store_data"
	syncEventID = "cached_store_sync"
)

// cachedStoreClusterEvent is a cluster event sent between nodes. It works for
// both the mutex-based and the single writer implementations. If Key is set,
// the message instructs to modify the key. If Data is nil, it is a delete
// operation. If IndexHash set, the receiver attempts to re-sync to it from the
// KV store. differentiate between put and delete (nil) events.
type cachedStoreClusterEvent[T any] struct {
	Value     *T     `json:"value,omitempty"`
	IndexHash string `json:"index_hash,omitempty"`
	Key       string `json:"key"`
	StoreName string `json:"name"`
}

func NewCachedStoreCluster(api config.API, kind CachedStoreClusterKind) *CachedStoreCluster {
	if kind == "" {
		kind = SingleWriterCachedStoreKind
	}
	return &CachedStoreCluster{
		api:              api,
		eventHandlers:    &sync.Map{},
		defaultStoreKind: kind,
	}
}

func (s *Service) OnPluginClusterEvent(r *incoming.Request, ev model.PluginClusterEvent) {
	f, err := s.cluster.getEventHandler(ev)
	if err != nil {
		r.Log.WithError(err).Errorw("failed to find a handler for plugin cluster event")
		return
	}
	err = f(r, ev)
	if err != nil {
		r.Log.WithError(err).Errorw("failed to handle plugin cluster event")
		return
	}
}

func (c *CachedStoreCluster) broadcastEvent(r *incoming.Request, id string, data any) {
	bb, err := json.Marshal(data)
	if err != nil {
		r.Log.WithError(err).Errorw("failed to marshal plugin cluster event")
		return
	}

	r = r.Clone()
	go func() {
		err = c.api.Plugin.PublishPluginClusterEvent(
			model.PluginClusterEvent{Id: id, Data: bb},
			model.PluginClusterEventSendOptions{SendType: model.PluginClusterEventSendTypeReliable},
		)
		if err != nil {
			r.Log.WithError(err).Errorw("failed to marshal plugin cluster event")
			return
		}
	}()
}

func (c *CachedStoreCluster) getEventHandler(ev model.PluginClusterEvent) (eventHandler, error) {
	v, ok := c.eventHandlers.Load(ev.Id)
	if !ok {
		return nil, utils.NewInvalidError("OnPluginClusterEvent: no handler for %s", ev.Id)
	}
	f, ok := v.(eventHandler)
	if !ok {
		return nil, utils.NewInvalidError("OnPluginClusterEvent: handler for %s is wrong type %T, expected %T", ev.Id, v, f)
	}
	return f, nil
}

func (c *CachedStoreCluster) setEventHandler(eventID string, h eventHandler) {
	c.eventHandlers.Store(eventID, h)
}
