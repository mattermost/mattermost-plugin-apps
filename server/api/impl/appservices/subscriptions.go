// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *AppServices) Subscribe(sub *apps.Subscription) error {
	return a.store.Sub().Save(sub)
}

func (a *AppServices) Unsubscribe(sub *apps.Subscription) error {
	return a.store.Sub().Delete(sub)
}
