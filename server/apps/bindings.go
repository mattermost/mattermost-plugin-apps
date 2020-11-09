package apps

import (
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

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *service) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	apps := s.ListApps()

	all := []*api.Binding{}
	for _, app := range apps {
		appCC := *cc
		appCC.AppID = app.Manifest.AppID
		bb, err := s.Client.GetBindings(&appCC)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get bindings for %s", app.Manifest.AppID)
		}

		scanned, err := s.scanAppBindings(app, bb, "")
		if err != nil {
			// TODO log the error!
			continue
		}

		all = mergeBindings(all, scanned)
	}

	return all, nil
}

// scanAppBindings removes bindings to locations that have not been granted to
// the App, and sets the AppID on the relevant elements.
func (s *service) scanAppBindings(app *api.App, bindings []*api.Binding, locPrefix api.Location) ([]*api.Binding, error) {
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
			return nil, errors.Errorf("location %s is not granted to app %s", fql, app.Manifest.AppID)
		}

		if !fql.IsTop() {
			b.AppID = app.Manifest.AppID
		}

		if len(b.Bindings) != 0 {
			scanned, err := s.scanAppBindings(app, b.Bindings, fql)
			if err != nil {
				return nil, err
			}
			b.Bindings = scanned
		}
	}

	return out, nil
}
