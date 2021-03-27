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
)

var ErrNotABot = errors.New("not a bot")
var ErrIsABot = errors.New("is a bot")

type Service interface {
	Subscribe(actingUserID string, _ *apps.Subscription) error
	Unsubscribe(actingUserID string, _ *apps.Subscription) error

	KVSet(botUserID, prefix, id string, ref interface{}) (bool, error)
	KVGet(botUserID, prefix, id string, ref interface{}) error
	KVDelete(botUserID, prefix, id string) error

	CreateOAuth2State(actingUserID string) (string, error)
	ValidateOAuth2State(actingUserID string, state string) error
	StoreRemoteOAuth2App(botUserID string, oapp apps.OAuth2App) error
	GetRemoteOAuth2User(_ apps.AppID, actingUserID string, ref interface{}) error
	StoreRemoteOAuth2User(_ apps.AppID, actingUserID string, ref interface{}) error
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
	mmuser, err := a.mm.User.Get(mattermostUserID)
	if err != nil {
		return err
	}
	if mmuser.IsBot {
		return errors.Wrap(ErrIsABot, mmuser.GetDisplayName(model.SHOW_NICKNAME_FULLNAME))
	}
	return nil
}
