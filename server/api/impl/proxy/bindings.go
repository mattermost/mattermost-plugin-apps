package proxy

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/pkg/errors"
)

func mergeBindings(bb1, bb2 []*api.Binding) []*api.Binding {
	out := append([]*api.Binding(nil), bb1...)

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

func (p *Proxy) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	allApps := p.store.ListApps()

	all := []*api.Binding{}
	for _, app := range allApps {
		appID := app.Manifest.AppID
		appCC := *cc
		appCC.AppID = appID

		up, err := p.upstreamForApp(app)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to make upstream for %s", appID)
		}

		// TODO PERF: Add caching
		// TODO PERF: Fan out the calls, wait for all to complete
		bindingsCall := app.Manifest.Bindings
		if bindingsCall == nil {
			bindingsCall = api.DefaultBindingsCall
		}
		bindingsCall.Context = cc

		bindings, err := up.GetBindings(bindingsCall)
		if err != nil {
			p.mm.Log.Error(fmt.Sprintf("failed to get bindings for %s: %v", appID, err))
		}
		all = mergeBindings(all, p.scanAppBindings(app, bindings, ""))
	}

	return all, nil
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (p *Proxy) scanAppBindings(app *api.App, bindings []*api.Binding, locPrefix api.Location) []*api.Binding {
	out := []*api.Binding{}
	for _, appB := range bindings {
		// clone just in case
		b := *appB
		fql := locPrefix.Make(b.Location)
		allowed := false
		for _, grantedLoc := range app.GrantedLocations {
			if fql.In(grantedLoc) || grantedLoc.In(fql) {
				allowed = true
				break
			}
		}
		if !allowed {
			// TODO Log this somehow to the app?
			p.mm.Log.Debug(fmt.Sprintf("location %s is not granted to app %s", fql, app.Manifest.AppID))
			continue
		}

		if !fql.IsTop() {
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
