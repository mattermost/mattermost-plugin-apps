package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *AppServices) KVSet(r *incoming.Request, botUserID, prefix, id string, data []byte) (bool, error) {
	if err := a.ensureFromBot(botUserID); err != nil {
		return false, err
	}
	return a.store.AppKV.Set(r, botUserID, prefix, id, data)
}

func (a *AppServices) KVGet(r *incoming.Request, botUserID, prefix, id string) ([]byte, error) {
	if err := a.ensureFromBot(botUserID); err != nil {
		return nil, err
	}
	return a.store.AppKV.Get(r, botUserID, prefix, id)
}

func (a *AppServices) KVDelete(r *incoming.Request, botUserID, prefix, id string) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}
	return a.store.AppKV.Delete(r, botUserID, prefix, id)
}

func (a *AppServices) KVList(r *incoming.Request, botUserID, prefix string, processf func(key string) error) error {
	if err := a.ensureFromBot(botUserID); err != nil {
		return err
	}
	return a.store.AppKV.List(r, botUserID, prefix, processf)
}
