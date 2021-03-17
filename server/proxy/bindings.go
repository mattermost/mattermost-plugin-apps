package proxy

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/apps"
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
func (p *Proxy) GetBindings(debugSessionToken apps.SessionToken, cc *apps.Context) ([]*apps.Binding, error) {
	allApps := p.store.App.AsMap()

	all := []*apps.Binding{}
	for _, app := range allApps {
		if !p.AppIsEnabled(app) {
			continue
		}
		appID := app.AppID
		appCC := *cc
		appCC.AppID = appID
		appCC.BotAccessToken = app.BotAccessToken

		// TODO PERF: Add caching
		// TODO PERF: Fan out the calls, wait for all to complete
		bindingsCall := apps.DefaultBindingsCall.WithOverrides(app.Bindings)
		bindingsRequest := &apps.CallRequest{
			Call:    *bindingsCall,
			Context: &appCC,
		}

		resp := p.Call(debugSessionToken, bindingsRequest)
		if resp == nil || resp.Type != apps.CallResponseTypeOK {
			// TODO Log error (chance to flood the logs)
			// p.mm.Log.Debug("Response is nil or unexpected type.")
			// if resp != nil && resp.Type == apps.CallResponseTypeError {
			// 	p.mm.Log.Debug("Error getting bindings. Error: " + resp.Error())
			// }
			continue
		}

		var bindings = []*apps.Binding{}
		b, _ := json.Marshal(resp.Data)
		err := json.Unmarshal(b, &bindings)
		if err != nil {
			// TODO Log error (chance to flood the logs)
			// p.mm.Log.Debug("Bindings are not of the right type.")
			continue
		}

		all = mergeBindings(all, p.scanAppBindings(app, bindings, ""))
	}

	return all, nil
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (p *Proxy) scanAppBindings(app *apps.App, bindings []*apps.Binding, locPrefix apps.Location) []*apps.Binding {
	out := []*apps.Binding{}
	locationsUsed := map[apps.Location]bool{}
	labelsUsed := map[string]bool{}

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
			// p.mm.Log.Debug(fmt.Sprintf("location %s is not granted to app %s", fql, app.Manifest.AppID))
			continue
		}

		if locPrefix == apps.LocationCommand {
			b.Location = apps.Location(app.Manifest.AppID)
			b.Label = string(app.Manifest.AppID)
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

		if len(b.Bindings) != 0 {
			scanned := p.scanAppBindings(app, b.Bindings, fql)
			if len(scanned) == 0 {
				// We do not add bindings without any valid sub-bindings
				continue
			}
			b.Bindings = scanned
		}

		out = append(out, &b)
	}

	return out
}
