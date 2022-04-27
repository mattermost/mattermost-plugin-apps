package appservices

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *AppServices) KVSet(_ *incoming.Request, appID apps.AppID, actingUserID, prefix, id string, data []byte) (bool, error) {
	if !json.Valid(data) {
		return false, utils.NewInvalidError("payload is no valid json")
	}

	return a.store.AppKV.Set(appID, actingUserID, prefix, id, data)
}

// KVGet returns the stored KV data for a given user and app.
// If err != nil, the returned data is always valid JSON.
func (a *AppServices) KVGet(_ *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) ([]byte, error) {
	data, err := a.store.AppKV.Get(appID, actingUserID, prefix, id)
	if err != nil && !errors.Is(err, utils.ErrNotFound) {
		return nil, err
	}

	if len(data) == 0 {
		// Ensure valid json is returned even if no data is set yet
		data = []byte(string("{}"))
	}

	return data, nil
}

func (a *AppServices) KVDelete(_ *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) error {
	return a.store.AppKV.Delete(appID, actingUserID, prefix, id)
}

func (a *AppServices) KVList(r *incoming.Request, appID apps.AppID, actingUserID, prefix string, processf func(key string) error) error {
	return a.store.AppKV.List(r, appID, actingUserID, prefix, processf)
}
