package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type KVStore struct{}

func (s *KVStore) Set(r *incoming.Request, prefix, id string, data []byte) (bool, error) {
	if r.SourceAppID() == "" || r.ActingUserID() == "" {
		return false, utils.NewInvalidError("source app ID or user ID missing in the request")
	}
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return false, err
	}

	set, err := r.Config().MattermostAPI().KV.Set(key, data)
	if err != nil {
		return false, err
	}
	if set {
		r.Log.Debugw("AppKV set", "prefix", prefix, "id", id, "hashkey", key)
	}
	return set, nil
}

func (s *KVStore) Get(r *incoming.Request, prefix, id string) ([]byte, error) {
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return nil, err
	}

	var data []byte
	if err = r.Config().MattermostAPI().KV.Get(key, &data); err != nil {
		return nil, err
	}

	return data, err
}

func (s *KVStore) Delete(r *incoming.Request, prefix, id string) error {
	key, err := Hashkey(KVAppPrefix, r.SourceAppID(), r.ActingUserID(), prefix, id)
	if err != nil {
		return err
	}

	err = r.Config().MattermostAPI().KV.Delete(key)
	if err != nil {
		return err
	}
	r.Log.Debugw("AppKV deleted", "prefix", prefix, "id", id, "hashkey", key)
	return nil
}

func (s *KVStore) List(r *incoming.Request, namespace string, processf func(key string) error) error {
	return ListHashKeys(r, processf,
		WithPrefix(KVAppPrefix),
		WithAppID(r.SourceAppID()),
		WithUserID(r.ActingUserID()),
		WithNamespace(namespace))
}
