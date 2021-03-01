// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (adm *Admin) GetApp(appID apps.AppID) (*apps.App, error) {
	return adm.store.App().Get(appID)
}

func (adm *Admin) GetManifest(appID apps.AppID) (*apps.Manifest, error) {
	return adm.store.Manifest().Get(appID)
}

func (adm *Admin) ListInstalledApps() map[apps.AppID]*apps.App {
	return adm.store.App().AsMap()
}

func (adm *Admin) ListMarketplaceApps(filter string) map[apps.AppID]*apps.MarketplaceApp {
	out := map[apps.AppID]*apps.MarketplaceApp{}

	for appID, m := range adm.store.Manifest().AsMap() {
		if !appMatchesFilter(m, filter) {
			continue
		}
		marketApp := &apps.MarketplaceApp{
			Manifest: m,
		}
		app, _ := adm.store.App().Get(appID)
		if app != nil {
			marketApp.Installed = true
			marketApp.Enabled = !app.Disabled
		}

		out[appID] = marketApp
	}

	return out
}

// Copied from Mattermost Server
func appMatchesFilter(manifest *apps.Manifest, filter string) bool {
	filter = strings.TrimSpace(strings.ToLower(filter))

	if filter == "" {
		return true
	}

	if strings.ToLower(string(manifest.AppID)) == filter {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.DisplayName), filter) {
		return true
	}

	if strings.Contains(strings.ToLower(manifest.Description), filter) {
		return true
	}

	return false
}
