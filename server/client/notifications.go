// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package client

import (
	"path"

	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
)

func (c *client) SendNotification(ss appmodel.SubscriptionSubject, msg interface{}) {
	c.DoPost(path.Join(c.app.Manifest.RootURL, "notify", string(ss)), msg)
}
