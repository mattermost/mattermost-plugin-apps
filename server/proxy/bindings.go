package proxy

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
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
func (p *Proxy) GetBindings(r *incoming.Request, cc apps.Context) ([]apps.Binding, error) {
	all := make(chan []apps.Binding)
	defer close(all)

	allApps := store.SortApps(p.store.App.AsMap(r))
	for i := range allApps {
		app := allApps[i]
		copy := r.Clone()
		copy.SetAppID(app.AppID)

		go func(r *incoming.Request, app apps.App) {
			bb := p.GetAppBindings(r, cc, app)
			all <- bb
		}(copy, app)
	}

	ret := []apps.Binding{}
	for i := 0; i < len(allApps); i++ {
		bb := <-all
		ret = mergeBindings(ret, bb)
	}
	return ret, nil
}

// GetAppBindings fetches bindings for a specific apps. We should avoid
// unnecessary logging here as this route is called very often.
func (p *Proxy) GetAppBindings(r *incoming.Request, cc apps.Context, app apps.App) []apps.Binding {
	if !p.appIsEnabled(r, app) {
		return nil
	}

	if len(app.GrantedLocations) == 0 {
		return nil
	}

	appID := app.AppID
	cc.AppID = appID

	// TODO PERF: Add caching
	bindingsCall := app.Bindings.WithDefault(apps.DefaultBindings)

	// no need to clean the context, Call will do.
	resp := p.call(r, app, bindingsCall, &cc)
	switch resp.Type {
	case apps.CallResponseTypeOK:
		var bindings = []apps.Binding{}
		b, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(b, &bindings)
		if err != nil {
			r.Log.WithError(err).Debugf("Bindings are not of the right type.")
			return nil
		}

		bindings = p.scanAppBindings(r, app, bindings, "", cc.UserAgent)
		return bindings

	case apps.CallResponseTypeError:
		r.Log.WithError(resp).Debugf("Error getting bindings")
		return nil

	default:
		r.Log.Debugf("Bindings response is nil or unexpected type.")
		return nil
	}
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (p *Proxy) scanAppBindings(r *incoming.Request, app apps.App, bindings []apps.Binding, locPrefix apps.Location, userAgent string) []apps.Binding {
	out := []apps.Binding{}
	locationsUsed := map[apps.Location]bool{}
	labelsUsed := map[string]bool{}

	conf := r.Config().Get()

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
			r.Log.Debugw("location is not granted to app", "location", fql)
			continue
		}

		if fql.In(apps.LocationCommand) {
			label := b.Label
			if label == "" {
				label = string(b.Location)
			}

			if strings.ContainsAny(label, " \t") {
				r.Log.Debugw("Binding validation error: Command label has multiple words", "location", b.Location)
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
				r.Log.WithError(err).Debugw("Invalid icon path in binding",
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
				r.Log.Debugw("Channel header button for webapp without icon", "label", b.Label)
				continue
			}
		}

		if len(b.Bindings) != 0 {
			scanned := p.scanAppBindings(r, app, b.Bindings, fql, userAgent)
			if len(scanned) == 0 {
				// We do not add bindings without any valid sub-bindings
				continue
			}
			b.Bindings = scanned
		}

		if b.Form != nil {
			clean, problems := cleanForm(*b.Form)
			for _, prob := range problems {
				r.Log.WithError(prob).Debugf("invalid form field in binding")
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
