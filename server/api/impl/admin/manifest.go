// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (adm *Admin) InstallManifest(cc *apps.Context, sessionToken apps.SessionToken, m *apps.Manifest) (md.MD, error) {
	if m.AppID == "" {
		return "", errors.New("app ID must not be empty")
	}

	// TODO check if acting user is a sysadmin

	adm.store.Manifest().StoreLocal(m)

	return md.Markdownf("Stored local manifest for %s [%s](%s).", m.AppID, m.DisplayName, m.HomepageURL), nil
}
