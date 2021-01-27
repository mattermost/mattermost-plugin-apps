// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *AppServices) Subscribe(sub *apps.Subscription) error {
	err := a.store.StoreSub(sub)
	if err != nil {
		return err
	}
	return nil
}

func (a *AppServices) Unsubscribe(sub *apps.Subscription) error {
	err := a.store.DeleteSub(sub)
	if err != nil {
		return err
	}
	return nil
}
