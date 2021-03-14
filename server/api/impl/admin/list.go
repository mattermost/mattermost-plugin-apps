// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package admin

import (
	"sort"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (adm *Admin) GetManifest(appID apps.AppID) (*apps.Manifest, error) {
	return adm.store.Manifest().Get(appID)
}

func (adm *Admin) GetInstalledApp(appID apps.AppID) (*apps.App, error) {
	return adm.store.App().Get(appID)
}

func (adm *Admin) GetInstalledApps() []*apps.App {
	installed := adm.store.App().AsMap()
	out := []*apps.App{}
	for _, app := range installed {
		out = append(out, app)
	}

	// Sort result alphabetically, byu display name.
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].DisplayName) < strings.ToLower(out[j].DisplayName)
	})

	return out
}

func (adm *Admin) GetListedApps(filter string) []*apps.ListedApp {
	out := []*apps.ListedApp{}

	for _, m := range adm.store.Manifest().AsMap() {
		if !appMatchesFilter(m, filter) {
			continue
		}
		marketApp := &apps.ListedApp{
			Manifest: m,
		}
		app, _ := adm.store.App().Get(m.AppID)
		if app != nil {
			marketApp.Installed = true
			marketApp.Enabled = !app.Disabled
		}
		out = append(out, marketApp)
	}

	// Sort result alphabetically, byu display name.
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Manifest.DisplayName) < strings.ToLower(out[j].Manifest.DisplayName)
	})

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
