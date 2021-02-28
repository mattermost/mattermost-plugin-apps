// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func (adm *Admin) ListApps() (map[apps.AppID]*apps.App, md.MD, error) {
	apps := adm.store.App().AsMap()
	if len(apps) == 0 {
		return nil, "no apps installed", nil
	}

	out := md.MD(`
| ID  | Type | OAuth2 | Bot | Locations | Permissions |
| :-- |:-----| :----- | :-- | :-------- | :---------- |
`)
	for _, app := range apps {
		out += md.Markdownf(`|%s|%s|%s|%s|%s|%s|
		`, app.AppID, app.Type, app.OAuth2ClientID, app.BotUserID, app.GrantedLocations, app.GrantedPermissions)
	}

	return apps, out, nil
}

func (adm *Admin) GetApp(appID apps.AppID) (*apps.App, error) {
	return adm.store.App().Get(appID)
}

func (adm *Admin) GetManifest(appID apps.AppID) (*apps.Manifest, error) {
	return adm.store.Manifest().Get(appID)
}
