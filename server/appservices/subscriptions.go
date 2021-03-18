// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

func (a *AppServices) Subscribe(actingUserID string, sub *apps.Subscription) error {
	err := utils.EnsureSysadmin(a.mm, actingUserID)
	if err != nil {
		return err
	}

	return a.store.Subscription.Save(sub)
}

func (a *AppServices) Unsubscribe(actingUserID string, sub *apps.Subscription) error {
	err := utils.EnsureSysadmin(a.mm, actingUserID)
	if err != nil {
		return err
	}

	return a.store.Subscription.Delete(sub)
}
