// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (p *Proxy) AddLocalManifest(cc *apps.Context, sessionToken apps.SessionToken, m *apps.Manifest) (md.MD, error) {
	if err := m.IsValid(); err != nil {
		return "", err
	}

	// TODO check if acting user is a sysadmin

	err := p.store.Manifest.StoreLocal(m)
	if err != nil {
		return "", err
	}

	return md.Markdownf("Stored local manifest for %s [%s](%s).", m.AppID, m.DisplayName, m.HomepageURL), nil
}
