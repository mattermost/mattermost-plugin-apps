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

	// Internal

	DeleteAppData(r *incoming.Request, appID apps.AppID, force bool) error
	RunCachedStoreTest(r *incoming.Request, s store.CachedStoreTestParams)
}

type AppServices struct {
	store *store.Service
}

var _ Service = (*AppServices)(nil)

func NewService(store *store.Service) *AppServices {
	return &AppServices{
		store: store,
	}
}

func (a *AppServices) DeleteAppData(r *incoming.Request, appID apps.AppID, force bool) (err error) {
	defer func() {
		if err != nil {
			r.Log.WithError(err).Errorf("Failed to clean up app data for %s: %v", appID, err)
		}
	}()

	if err = r.Check(
		r.RequireActingUser,
		r.RequireSysadminOrPlugin,
	); err != nil {
		return err
	}

	// Remove all KV and user data.
	if err = store.RemoveAllKVAndUserDataForApp(r, appID); err != nil {
		return err
	}

	// Remove all subscriptions.
	if err = a.unsubscribeApp(r, appID); err != nil {
		return err
	}
	return nil
}
