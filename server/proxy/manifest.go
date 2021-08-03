// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) AddLocalManifest(m apps.Manifest) (string, error) {
	if err := m.IsValid(); err != nil {
		return "", err
	}

	err := p.store.Manifest.StoreLocal(m)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Stored local manifest for %s [%s](%s).", m.AppID, m.DisplayName, m.HomepageURL), nil
}
