// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"errors"

	"github.com/mattermost/mattermost-plugin-apps/modelapps"
)

var ErrNotABot = errors.New("not a bot")

type AppServices interface {
	Subscribe(*modelapps.Subscription) error
	Unsubscribe(*modelapps.Subscription) error
	KVSet(botUserID, prefix, id string, ref interface{}) (bool, error)
	KVGet(botUserID, prefix, id string, ref interface{}) error
	KVDelete(botUserID, prefix, id string) error
}
