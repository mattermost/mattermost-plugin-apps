// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *AppServices) Subscribe(actingUserID string, sub *apps.Subscription) error {
	err := utils.EnsureSysAdmin(a.conf.MattermostAPI(), actingUserID)
	if err != nil {
		return err
	}

	return a.store.Subscription.Save(sub)
}

func (a *AppServices) GetSubscriptions(actingUserID string) ([]*apps.Subscription, error) {
	err := utils.EnsureSysAdmin(a.mm, actingUserID)
	if err != nil {
		return nil, err
	}

	return a.store.Subscription.List()
}

func (a *AppServices) Unsubscribe(actingUserID string, sub *apps.Subscription) error {
	err := utils.EnsureSysAdmin(a.conf.MattermostAPI(), actingUserID)
	if err != nil {
		return err
	}

	return a.store.Subscription.Delete(sub)
}
