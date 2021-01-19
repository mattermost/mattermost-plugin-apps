// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appservices

import (
	"github.com/mattermost/mattermost-plugin-apps/modelapps"
)

func (a *AppServices) Subscribe(sub *modelapps.Subscription) error {
	return a.store.StoreSub(sub)
}

func (a *AppServices) Unsubscribe(sub *modelapps.Subscription) error {
	return a.store.DeleteSub(sub)
}
