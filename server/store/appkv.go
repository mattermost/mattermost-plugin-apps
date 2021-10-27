package store

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

const (
	keysPerPage = 1000
)

type AppKVStore interface {
	Set(botUserID, prefix, id string, ref interface{}) (bool, error)
	Get(botUserID, prefix, id string, ref interface{}) error
	Delete(botUserID, prefix, id string) error
	List(botUserID, namespace string, processf func(key string) error) error
}

type appKVStore struct {
	*Service
}

var _ AppKVStore = (*appKVStore)(nil)

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API
func (s *appKVStore) Set(botUserID, prefix, id string, ref interface{}) (bool, error) {
	key, err := Hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return false, err
	}
	return s.conf.MattermostAPI().KV.Set(key, ref)
}

func (s *appKVStore) Get(botUserID, prefix, id string, ref interface{}) error {
	key, err := Hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return err
	}
	return s.conf.MattermostAPI().KV.Get(key, ref)
}

func (s *appKVStore) Delete(botUserID, prefix, id string) error {
	key, err := Hashkey(config.KVAppPrefix, botUserID, prefix, id)
	if err != nil {
		return err
	}
	return s.conf.MattermostAPI().KV.Delete(key)
}

func (s *appKVStore) List(
	botUserID, namespace string,
	processf func(key string) error,
) error {
	_, mm, log := s.conf.Basic()
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

			_, _, ns, _, err := ParseHashkey(key)
			if err != nil {
				log.WithError(err).Debugw("failed to parse key", "key", key)
				continue
			}
			if namespace != "" && ns != namespace {
				continue
			}

			err = processf(key)
		}

		if len(keys) < keysPerPage {
			break
		}
	}
	return nil
}
