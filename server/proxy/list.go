// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"sort"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (p *Proxy) GetManifest(appID apps.AppID) (*apps.Manifest, error) {
	return p.store.Manifest.Get(appID)
}

func (p *Proxy) GetManifestFromS3(appID apps.AppID, version apps.AppVersion) (*apps.Manifest, error) {
	return p.store.Manifest.GetFromS3(appID, version)
}

func (p *Proxy) GetInstalledApp(appID apps.AppID) (*apps.App, error) {
	return p.store.App.Get(appID)
}

func (p *Proxy) GetInstalledApps() []*apps.App {
	installed := p.store.App.AsMap()
	out := []*apps.App{}
	for _, app := range installed {
		out = append(out, app)
	}

	// Sort result alphabetically, by display name.
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].DisplayName) < strings.ToLower(out[j].DisplayName)
	})

	return out
}

func (p *Proxy) GetListedApps(filter string) []*apps.ListedApp {
	out := []*apps.ListedApp{}

	for _, m := range p.store.Manifest.AsMap() {
		if !appMatchesFilter(m, filter) {
			continue
		}
		marketApp := &apps.ListedApp{
			Manifest: m,
		}
		app, _ := p.store.App.Get(m.AppID)
		if app != nil {
			marketApp.Installed = true
			marketApp.Enabled = !app.Disabled
			marketApp.Labels = []model.MarketplaceLabel{{
				Name:        "Experimental",
				Description: "Apps are marked as experimental and not meant for production use. Please use with caution.",
				URL:         "",
			}}
		}
		out = append(out, marketApp)
	}

	// Sort result alphabetically, by display name.
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
