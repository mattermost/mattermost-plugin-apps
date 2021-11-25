package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *AppServices) KVSet(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string, data []byte) (bool, error) {
	return a.store.AppKV.Set(r, appID, actingUserID, prefix, id, data)
}

func (a *AppServices) KVGet(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) ([]byte, error) {
	return a.store.AppKV.Get(r, appID, actingUserID, prefix, id)
}

func (a *AppServices) KVDelete(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) error {
	return a.store.AppKV.Delete(r, appID, actingUserID, prefix, id)
}

func (a *AppServices) KVList(r *incoming.Request, appID apps.AppID, actingUserID, prefix string, processf func(key string) error) error {
	return a.store.AppKV.List(r, appID, actingUserID, prefix, processf)
}
