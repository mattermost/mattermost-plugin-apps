package proxy

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func mergeBindings(bb1, bb2 []apps.Binding) []apps.Binding {
	out := append([]apps.Binding(nil), bb1...)

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
func (p *Proxy) GetBindings(in Incoming, cc apps.Context) ([]apps.Binding, error) {
	all := make(chan []apps.Binding)
	defer close(all)

	allApps := store.SortApps(p.store.App.AsMap())
	for i := range allApps {
		app := allApps[i]

		go func(app *apps.App) {
			bb := p.getBindingsForApp(in, cc, app)
			all <- bb
		}(&app)
	}

	ret := []apps.Binding{}
	for i := 0; i < len(allApps); i++ {
		bb := <-all
		ret = mergeBindings(ret, bb)
	}
	return ret, nil
}

// getBindingsForApp fetches bindings for a specific apps. We should avoid
// unnecessary logging here as this route is called very often.
func (p *Proxy) getBindingsForApp(in Incoming, cc apps.Context, app *apps.App) []apps.Binding {
	if !p.appIsEnabled(app) {
		return nil
	}
	log := p.conf.Logger().With("app_id", app.AppID)

	appID := app.AppID
	cc.AppID = appID

	// TODO PERF: Add caching
	bindingsCall := app.Bindings.WithDefault(apps.DefaultBindings)
	bindingsRequest := apps.CallRequest{
		Call: bindingsCall,
		// no need to clean the context, Call will do.
		Context: cc,
	}

	resp := p.callApp(in, app, bindingsRequest)
	if resp.Type != apps.CallResponseTypeError && resp.Type != apps.CallResponseTypeOK {
		log.Debugf("Bindings response is nil or unexpected type.")
		return nil
	}

	// TODO: ignore a 404, no bindings
	if resp.Type == apps.CallResponseTypeError {
		log.WithError(resp).Debugf("Error getting bindings")
		return nil
	}

	var bindings = []apps.Binding{}
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
func (p *Proxy) scanAppBindings(app *apps.App, bindings []apps.Binding, locPrefix apps.Location, userAgent string) []apps.Binding {
	out := []apps.Binding{}
	locationsUsed := map[apps.Location]bool{}
	labelsUsed := map[string]bool{}
	conf, _, log := p.conf.Basic()
	log = log.With("app_id", app.AppID)

	for _, b := range bindings {
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
			log.Debugw("location is not granted to app", "location", fql)
			continue
		}

		if fql.In(apps.LocationCommand) {
			label := b.Label
			if label == "" {
				label = string(b.Location)
			}

			if strings.ContainsAny(label, " \t") {
				log.Debugw("Binding validation error: Command label has multiple words", "location", b.Location)
				continue
			}
		}

		if fql.IsTop() {
			if locationsUsed[b.Location] {
				continue
			}
			locationsUsed[b.Location] = true
		} else {
			if b.Location == "" || b.Label == "" {
				continue
			}
			if locationsUsed[b.Location] || labelsUsed[b.Label] {
				continue
			}

			locationsUsed[b.Location] = true
			labelsUsed[b.Label] = true
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
				log.Debugw("Channel header button for webapp without icon", "label", b.Label)
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

		if b.Form != nil {
			clean, problems := cleanForm(*b.Form)
			for _, prob := range problems {
				log.WithError(prob).Debugf("invalid form field in bingding")
			}
			b.Form = &clean
		}

		out = append(out, b)
	}

	return out
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}
