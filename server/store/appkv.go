package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	keysPerPage = 1000
)

type AppKVStore interface {
	Set(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string, data []byte) (bool, error)
	Get(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) ([]byte, error)
	Delete(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) error
	List(r *incoming.Request, appID apps.AppID, actingUserID, namespace string, processf func(key string) error) error
}

type appKVStore struct {
	*Service
}

var _ AppKVStore = (*appKVStore)(nil)

func (s *appKVStore) Set(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string, data []byte) (bool, error) {
	if appID == "" || actingUserID == "" {
		return false, utils.NewInvalidError("app and user IDs must be provided")
	}

	key, err := Hashkey(config.KVAppPrefix, appID, actingUserID, prefix, id)
	if err != nil {
		return false, err
	}

	return s.conf.MattermostAPI().KV.Set(key, data)
}

func (s *appKVStore) Get(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) ([]byte, error) {
	key, err := Hashkey(config.KVAppPrefix, appID, actingUserID, prefix, id)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = s.conf.MattermostAPI().KV.Get(key, &data); err != nil {
		return nil, err
	}

	return data, err
}

func (s *appKVStore) Delete(r *incoming.Request, appID apps.AppID, actingUserID, prefix, id string) error {
	key, err := Hashkey(config.KVAppPrefix, appID, actingUserID, prefix, id)
	if err != nil {
		return err
	}

	return s.conf.MattermostAPI().KV.Delete(key)
}

func (s *appKVStore) List(
	r *incoming.Request,
	appID apps.AppID, actingUserID,
	namespace string, processf func(key string) error,
) error {
	mm := s.conf.MattermostAPI()
	for i := 0; ; i++ {
		keys, err := mm.KV.ListKeys(i, keysPerPage)
		if err != nil {
			return errors.Wrapf(err, "failed to list keys - page, %d", i)
		}

		for _, key := range keys {
			// all apps keys are 50 bytes
			if !strings.HasPrefix(key, config.KVAppPrefix) || len(key) != 50 {
				continue
			}

			_, _, _, ns, _, err := ParseHashkey(key)
			if err != nil {
				r.Log.WithError(err).Debugw("failed to parse key", "key", key)
				continue
			}
			if namespace != "" && ns != namespace {
				continue
			}

			err = processf(key)
			if err != nil {
				return err
			}
		}

		if len(keys) < keysPerPage {
			break
		}
	}
	return nil
}
