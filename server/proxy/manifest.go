// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (p *Proxy) AddLocalManifest(actingUserID string, sessionToken apps.SessionToken, m *apps.Manifest) (md.MD, error) {
	if err := m.IsValid(); err != nil {
		return "", err
	}

	err := utils.EnsureSysadmin(p.mm, actingUserID)
	if err != nil {
		return "", err
	}

	err = p.store.Manifest.StoreLocal(m)
	if err != nil {
		return "", err
	}

	return md.Markdownf("Stored local manifest for %s [%s](%s).", m.AppID, m.DisplayName, m.HomepageURL), nil
}
