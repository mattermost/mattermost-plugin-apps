// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type Service interface {
	// Subscriptions

	Subscribe(*incoming.Request, apps.Subscription) error
	GetSubscriptions(*incoming.Request) ([]apps.Subscription, error)
	Unsubscribe(*incoming.Request, apps.Event) error
	UnsubscribeApp(*incoming.Request, apps.AppID) error

	// KV

	KVSet(_ *incoming.Request, prefix, id string, data []byte) (bool, error)
	KVGet(_ *incoming.Request, prefix, id string) ([]byte, error)
	KVDelete(_ *incoming.Request, prefix, id string) error
	KVList(_ *incoming.Request, namespace string, processf func(key string) error) error
	KVDebugInfo(*incoming.Request) (*store.KVDebugInfo, error)
	KVDebugAppInfo(*incoming.Request, apps.AppID) (*store.KVDebugAppInfo, error)

	// Remote (3rd party) OAuth2

	StoreOAuth2App(_ *incoming.Request, data []byte) error
	StoreOAuth2User(_ *incoming.Request, data []byte) error
	GetOAuth2User(_ *incoming.Request) ([]byte, error)
}

type AppServices struct {
	apps          store.Apps
	kv            *store.KVStore
	oauth2        *store.OAuth2Store
	subscriptions *store.SubscriptionStore
}

var _ Service = (*AppServices)(nil)

func NewService(appStore store.Apps, kvStore *store.KVStore, oauth2Store *store.OAuth2Store) *AppServices {
	return &AppServices{
		apps:   appStore,
		kv:     kvStore,
		oauth2: oauth2Store,
	}
}
