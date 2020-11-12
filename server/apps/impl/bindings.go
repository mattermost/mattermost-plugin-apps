package impl

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/pkg/errors"
)

func mergeBindings(bb1, bb2 []*apps.Binding) []*apps.Binding {
	out := append([]*apps.Binding(nil), bb1...)

	for _, b2 := range bb2 {
		found := false
		for i, o := range out {
			if b2.AppID == o.AppID && b2.LocationID == o.LocationID {
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

func setAppID(bb []*apps.Binding, appID apps.AppID, excludeTopLevel bool) {
	for _, b := range bb {
		if !excludeTopLevel {
			b.AppID = appID
		}
		if len(b.Bindings) != 0 {
			setAppID(b.Bindings, appID, false)
		}
	}
}

// This and registry related calls should be RPC calls so they can be reused by other plugins
func (s *service) GetBindings(cc *apps.Context) ([]*apps.Binding, error) {
	appIDs, err := s.Store.ListApps()
	if err != nil {
		return nil, errors.Wrap(err, "error getting all app IDs")
	}

	all := []*apps.Binding{}
	for _, appID := range appIDs {
		appCC := *cc
		appCC.AppID = appID
		bb, err := s.Client.GetBindings(&appCC)
		if err != nil {
			s.Mattermost.Log.Warn("Failed to get bindings", "app_id", appID)
			continue
		}

		// TODO eliminate redundant AppID, just need it at the top level? I.e.
		// group by AppID instead of top-level LocationID
		setAppID(bb, appID, true)

		all = mergeBindings(all, bb)
	}

	return all, nil
}
