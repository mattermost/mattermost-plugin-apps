package proxy

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func mergeBindings(bb1, bb2 []*apps.Binding) []*apps.Binding {
	out := append([]*apps.Binding(nil), bb1...)

	for _, b2 := range bb2 {
		found := false
		for i, o := range out {
			if b2.AppID == o.AppID && b2.Location == o.Location {
				found = true

				// b2 overrides b1, if b1 and b2 have Bindings, they are merged
				merged := b2
				if len(o.Bindings) != 0 && b2.Call == nil {
					merged.Bindings = mergeBindings(o.Bindings, b2.Bindings)
				}
				out[i] = merged
			}
		}
		if !found {
			out = append(out, b2)
		}
	}
	return out
}

// GetBindings fetches bindings for all apps.
// We should avoid unnecessary logging here as this route is called very often.
func (p *Proxy) GetBindings(sessionID, actingUserID string, cc *apps.Context) ([]*apps.Binding, error) {
	allApps := store.SortApps(p.store.App.AsMap())
	all := make([][]*apps.Binding, len(allApps))

	var wg sync.WaitGroup
	for i, app := range allApps {
		wg.Add(1)
		go func(app *apps.App, i int) {
			defer wg.Done()
			all[i] = p.GetBindingsForApp(sessionID, actingUserID, cc, app)
		}(app, i)
	}
	wg.Wait()

	ret := []*apps.Binding{}
	for _, b := range all {
		ret = mergeBindings(ret, b)
	}

	return ret, nil
}

// GetBindingsForApp fetches bindings for a specific apps.
// We should avoid unnecessary logging here as this route is called very often.
func (p *Proxy) GetBindingsForApp(sessionID, actingUserID string, cc *apps.Context, app *apps.App) []*apps.Binding {
	if !p.AppIsEnabled(app) {
		return nil
	}
	log := p.conf.Logger().With("app_id", app.AppID)

	appID := app.AppID
	appCC := *cc
	appCC.AppID = appID
	appCC.BotAccessToken = app.BotAccessToken

	// TODO PERF: Add caching
	bindingsCall := apps.DefaultBindings.WithOverrides(app.Bindings)
	bindingsRequest := &apps.CallRequest{
		Call:    *bindingsCall,
		Context: &appCC,
	}

	resp := p.Call(sessionID, actingUserID, bindingsRequest)
	if resp == nil || (resp.Type != apps.CallResponseTypeError && resp.Type != apps.CallResponseTypeOK) {
		log.Debugf("Bindings response is nil or unexpected type.")
		return nil
	}

	// TODO: ignore a 404, no bindings
	if resp.Type == apps.CallResponseTypeError {
		log.WithError(resp).Debugf("Error getting bindings.")
		return nil
	}

	var bindings = []*apps.Binding{}
	b, _ := json.Marshal(resp.Data)
	err := json.Unmarshal(b, &bindings)
	if err != nil {
		log.Debugf("Bindings are not of the right type.")
		return nil
	}

	bindings = p.scanAppBindings(app, bindings, "", cc.UserAgent)

	return bindings
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (p *Proxy) scanAppBindings(app *apps.App, bindings []*apps.Binding, locPrefix apps.Location, userAgent string) []*apps.Binding {
	out := []*apps.Binding{}
	locationsUsed := map[apps.Location]bool{}
	labelsUsed := map[string]bool{}
	conf, _, log := p.conf.Basic()
	log = log.With("app_id", app.AppID)

	for _, appB := range bindings {
		// clone just in case
		b := *appB
		if b.Location == "" {
			b.Location = apps.Location(app.Manifest.AppID)
		}

		fql := locPrefix.Make(b.Location)
		allowed := false
		for _, grantedLoc := range app.GrantedLocations {
			if fql.In(grantedLoc) || grantedLoc.In(fql) {
				allowed = true
				break
			}
		}
		if !allowed {
			p.conf.MattermostAPI().Log.Debug("location is not granted to app", "location", fql, "appID", app.Manifest.AppID)
			continue
		}

		if fql.In(apps.LocationCommand) {
			label := b.Label
			if label == "" {
				label = string(b.Location)
			}

			if strings.ContainsAny(label, " \t") {
				p.conf.MattermostAPI().Log.Debug("Binding validation error: Command label has multiple words", "app", app.Manifest.AppID, "location", b.Location)
				continue
			}
		}

		if fql.IsTop() {
			if locationsUsed[appB.Location] {
				continue
			}
			locationsUsed[appB.Location] = true
		} else {
			if b.Location == "" || b.Label == "" {
				continue
			}
			if locationsUsed[appB.Location] || labelsUsed[appB.Label] {
				continue
			}

			locationsUsed[appB.Location] = true
			labelsUsed[appB.Label] = true
			b.AppID = app.Manifest.AppID
		}

		if b.Icon != "" {
			icon, err := normalizeStaticPath(conf, app.AppID, b.Icon)
			if err != nil {
				log.WithError(err).Debugw("Invalid icon path in binding",
					"app_id", app.AppID,
					"icon", b.Icon)
				b.Icon = ""
			} else {
				b.Icon = icon
			}
		}

		// First level of Channel Header
		if fql == apps.LocationChannelHeader.Make(b.Location) {
			// Must have an icon on webapp to show the icon
			if b.Icon == "" && userAgent == "webapp" {
				p.conf.MattermostAPI().Log.Debug("Channel header button for webapp without icon", "label", b.Label, "app_id", app.AppID)
				continue
			}
		}

		if len(b.Bindings) != 0 {
			scanned := p.scanAppBindings(app, b.Bindings, fql, userAgent)
			if len(scanned) == 0 {
				// We do not add bindings without any valid sub-bindings
				continue
			}
			b.Bindings = scanned
		}

		p.cleanForm(b.Form)

		out = append(out, &b)
	}

	return out
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}
