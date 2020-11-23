package impl

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/pkg/errors"
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

func addAppID(bb []*apps.Binding, appID apps.AppID) []*apps.Binding {
	for _, b := range bb {
		b.AppID = appID
		if len(b.Bindings) > 0 {
			b.Bindings = addAppID(b.Bindings, appID)
		}
	}

	return bb
}

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *service) GetBindings(cc *apps.Context) ([]*apps.Binding, error) {
	allApps := s.ListApps()

	all := []*apps.Binding{}
	for _, app := range allApps {
		appCC := *cc
		appCC.AppID = app.Manifest.AppID
		bb, err := s.Client.GetBindings(&appCC)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bindings for %s", app.Manifest.AppID)
		}

		bb = addAppID(bb, app.Manifest.AppID)
		all = mergeBindings(all, s.scanAppBindings(app, bb, ""))
	}

	return all, nil
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (s *service) scanAppBindings(app *apps.App, bindings []*apps.Binding, locPrefix apps.Location) []*apps.Binding {
	out := []*apps.Binding{}
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
			s.Mattermost.Log.Debug(fmt.Sprintf("location %s is not granted to app %s", fql, app.Manifest.AppID))
			continue
		}

		if !fql.IsTop() {
			b.AppID = app.Manifest.AppID
		}

		if len(b.Bindings) != 0 {
			scanned := s.scanAppBindings(app, b.Bindings, fql)
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
