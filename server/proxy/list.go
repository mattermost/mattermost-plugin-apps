// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

const pingAppTimeout = 1 * time.Second

func (p *Proxy) GetManifest(appID apps.AppID) (*apps.Manifest, error) {
	return p.manifestStore.Get(appID)
}

func (p *Proxy) GetApp(r *incoming.Request) (*apps.App, error) {
	if err := r.Check(
		r.RequireActingUser,
	); err != nil {
		return nil, err
	}

	app, err := p.appStore.Get(r.Destination())
	if err != nil {
		return nil, err
	}

	if err = r.RequireSysadminOrPlugin(); err != nil {
		// Sanitize for non-sysadmins.
		app.WebhookSecret = ""
		app.MattermostOAuth2 = nil
		app.RemoteOAuth2 = apps.OAuth2App{}
	}
	return app, nil
}

func (p *Proxy) GetInstalledApp(appID apps.AppID, checkEnabled bool) (*apps.App, error) {
	app, err := p.appStore.Get(appID)
	if err != nil {
		return nil, err
	}
	if checkEnabled {
		if err = p.ensureEnabled(app); err != nil {
			return nil, err
		}
	}
	return app, nil
}

func (p *Proxy) PingInstalledApps(ctx context.Context) (installed []apps.App, reachable map[apps.AppID]bool) {
	all := p.appStore.AsMap(store.AllApps)
	if len(all) == 0 {
		return nil, nil
	}

	// all ping requests must respond, unreachable respond with "".
	reachableCh := make(chan apps.AppID)
	defer close(reachableCh)
	for _, app := range all {
		go func(a apps.App) {
			var reachable bool

			if a.DeployType == apps.DeployBuiltin {
				// Builtin apps are always rechable
				reachable = true
			} else if !a.Disabled {
				if p.pingApp(ctx, &a) {
					reachable = true
				}
			}
			if reachable {
				reachableCh <- a.AppID
			} else {
				reachableCh <- ""
			}
		}(app)
	}

	for _, app := range all {
		installed = append(installed, app)
		appID := <-reachableCh
		if appID != "" {
			if reachable == nil {
				reachable = map[apps.AppID]bool{}
			}
			reachable[appID] = true
		}
	}

	// Sort result alphabetically, by display name.
	sort.SliceStable(installed, func(i, j int) bool {
		return strings.ToLower(installed[i].DisplayName) < strings.ToLower(installed[j].DisplayName)
	})

	return installed, reachable
}

func (p *Proxy) GetInstalledApps() []apps.App {
	return p.appStore.AsList(store.AllApps)
}

func (p *Proxy) GetListedApps(filter string, includePluginApps bool) []apps.ListedApp {
	conf := p.conf.Get()
	out := []apps.ListedApp{}

	for _, m := range p.manifestStore.AsMap() {
		if !appMatchesFilter(m, filter) {
			continue
		}

		marketApp := apps.ListedApp{
			Manifest: m,
		}

		if m.Icon != "" {
			marketApp.IconURL = conf.StaticURL(m.AppID, m.Icon)
		}

		app, _ := p.appStore.Get(m.AppID)

		if !includePluginApps {
			// Filter out if installed as plugin
			if app != nil && app.DeployType == apps.DeployPlugin {
				continue
			}

			// Filter out if not installed and only deployable as plugin
			if app == nil && len(m.DeployTypes()) == 1 && m.DeployTypes()[0] == apps.DeployPlugin {
				continue
			}
		}

		if app != nil {
			marketApp.Installed = true
			marketApp.Enabled = !app.Disabled

			if !marketApp.Enabled {
				marketApp.Labels = append(marketApp.Labels, model.MarketplaceLabel{
					Name:        "Disabled",
					Description: "This app is disabled.",
					URL:         "",
				})
			}
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
func appMatchesFilter(manifest apps.Manifest, filter string) bool {
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
