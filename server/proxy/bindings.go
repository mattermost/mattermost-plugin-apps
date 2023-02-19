package proxy

import (
	"sort"
	"time"

	"github.com/hashicorp/go-multierror"
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
				if len(o.Bindings) != 0 {
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
func (p *Proxy) GetBindings(r *incoming.Request, cc apps.Context) (ret []apps.Binding, err error) {
	start := time.Now()
	var allApps []apps.App
	defer func() {
		log := r.Log.With("elapsed", time.Since(start).String())
		if err != nil {
			log.WithError(err).Warnf("GetBindings failed")
		} else {
			log.Debugf("GetBindings: returned bindings for %v apps", len(allApps))
		}
	}()

	if err := r.Check(
		r.RequireActingUser,
	); err != nil {
		return nil, err
	}

	type result struct {
		appID    apps.AppID
		bindings []apps.Binding
		err      error
	}

	all := make(chan result)
	defer close(all)

	allApps = p.store.App.AsList(store.EnabledAppsOnly)
	for i := range allApps {
		go func(app apps.App) {
			apprequest := r.WithDestination(app.AppID)
			res := result{
				appID: app.AppID,
			}
			res.bindings, res.err = p.InvokeGetBindings(apprequest, cc)
			if res.err != nil {
				r.Log.WithError(res.err).Debugf("failed to fetch app bindings")
			}
			all <- res
		}(allApps[i])
	}

	ret = []apps.Binding{}
	var problems error
	for i := 0; i < len(allApps); i++ {
		res := <-all
		ret = mergeBindings(ret, res.bindings)
		if res.err != nil {
			problems = multierror.Append(problems, res.err)
		}
	}

	return SortTopBindings(ret), problems
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}

// SortTopBindings ensures that the top-level bindings are sorted by Location,
// and their sub-bindings are sorted by their AppID. The children of those are
// left untouched.
func SortTopBindings(in []apps.Binding) (out []apps.Binding) {
	for _, b := range in {
		sort.Slice(b.Bindings, func(i, j int) bool { return b.Bindings[i].AppID < b.Bindings[j].AppID })
		out = append(out, b)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Location < out[j].Location })
	return out
}
