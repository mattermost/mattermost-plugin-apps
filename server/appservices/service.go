// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

var ErrNotABot = errors.New("not a bot")

type Service interface {
	Subscribe(*apps.Subscription) error
	Unsubscribe(*apps.Subscription) error
	KVSet(botUserID, prefix, id string, ref interface{}) (bool, error)
	KVGet(botUserID, prefix, id string, ref interface{}) error
	KVDelete(botUserID, prefix, id string) error
}

type AppServices struct {
	mm    *pluginapi.Client
	conf  config.Service
	store *store.Service
}

var _ Service = (*AppServices)(nil)

func NewService(mm *pluginapi.Client, conf config.Service, store *store.Service) *AppServices {
	return &AppServices{
		mm:    mm,
		conf:  conf,
		store: store,
	}
}
