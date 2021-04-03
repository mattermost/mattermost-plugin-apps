// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

var ErrNotABot = errors.New("not a bot")
var ErrIsABot = errors.New("is a bot")

type Service interface {
	// Subscriptions

	Subscribe(actingUserID string, _ *apps.Subscription) error
	Unsubscribe(actingUserID string, _ *apps.Subscription) error

	// KV

	// ref can be either a []byte for raw data, or anything else will be JSON marshaled.
	KVSet(botUserID, prefix, id string, ref interface{}) (bool, error)
	KVGet(botUserID, prefix, id string, ref interface{}) error
	KVDelete(botUserID, prefix, id string) error

	// Remote (3rd party) OAuth2

	StoreOAuth2App(_ apps.AppID, actingUserID string, oapp apps.OAuth2App) error
	GetOAuth2User(_ apps.AppID, actingUserID string, ref interface{}) error
	// ref can be either a []byte, or anything else will be JSON marshaled.
	StoreOAuth2User(_ apps.AppID, actingUserID string, ref interface{}) error
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

func (a *AppServices) ensureFromBot(mattermostUserID string) error {
	if mattermostUserID == "" {
		return utils.NewUnauthorizedError("not logged in")
	}
	mmuser, err := a.mm.User.Get(mattermostUserID)
	if err != nil {
		return err
	}
	if !mmuser.IsBot {
		return errors.Wrap(ErrNotABot, mmuser.GetDisplayName(model.SHOW_NICKNAME_FULLNAME))
	}
	return nil
}

func (a *AppServices) ensureFromUser(mattermostUserID string) error {
	if mattermostUserID == "" {
		return utils.NewUnauthorizedError("not logged in")
	}
	mmuser, err := a.mm.User.Get(mattermostUserID)
	if err != nil {
		return err
	}
	if mmuser.IsBot {
		return errors.Wrap(ErrIsABot, mmuser.GetDisplayName(model.SHOW_NICKNAME_FULLNAME))
	}
	return nil
}
