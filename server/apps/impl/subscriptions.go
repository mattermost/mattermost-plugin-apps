// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package impl

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (s *service) Subscribe(sub *apps.Subscription) error {
	return s.store.StoreSub(sub)
}

func (s *service) Unsubscribe(sub *apps.Subscription) error {
	return s.store.DeleteSub(sub)
}
