package appservices

import (
	"crypto/md5" // nolint:gosec
	"unicode/utf8"

	"encoding/hex"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API
func (a *AppServices) KVSet(botUserID, prefix, id string, ref interface{}) (bool, error) {
	changed := false
	var err error
	err = a.kvDo(botUserID, prefix, id, ref, func(key string, ref interface{}) error {
		changed, err = a.mm.KV.Set(key, ref)
		return err
	})
	if err != nil {
		return false, err
	}
	return changed, nil
}

func (a *AppServices) KVGet(botUserID, prefix, id string, ref interface{}) error {
	return a.kvDo(botUserID, prefix, id, ref, a.mm.KV.Get)
}

func (a *AppServices) KVDelete(botUserID, prefix, id string) error {
	return a.kvDo(botUserID, prefix, id, nil, func(key string, _ interface{}) error {
		return a.mm.KV.Delete(key)
	})
}

func kvKey(namespace, prefix, id string) string {
	if id == "" || namespace == "" {
		return ""
	}

	namespaceHash := md5.Sum([]byte(namespace)) // nolint:gosec
	idHash := md5.Sum([]byte(id))               // nolint:gosec
	key := strings.Join([]string{
		hex.EncodeToString(namespaceHash[:]),
		prefix,
		hex.EncodeToString(idHash[:]),
	}, "/")

	if utf8.RuneCountInString(key) > model.KEY_VALUE_KEY_MAX_RUNES {
		return key[:model.KEY_VALUE_KEY_MAX_RUNES]
	}

	return key
}

// TODO use raw byte API: for now all JSON is re-encoded to use api.Mattermost API
func (a *AppServices) kvDo(botUserID, prefix, id string, ref interface{}, f func(key string, ref interface{}) error) error {
	mmuser, err := a.mm.User.Get(botUserID)
	if err != nil {
		return err
	}

	if !mmuser.IsBot {
		return api.ErrNotABot
	}

	return f(kvKey(botUserID, prefix, id), ref)
}
