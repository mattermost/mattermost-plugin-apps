package proxy

import (
	"github.com/hashicorp/go-multierror"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pkg/errors"

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
func (p *Proxy) GetBindings(r *incoming.Request, cc apps.Context) ([]apps.Binding, error) {
	type result struct {
		appID    apps.AppID
		bindings []apps.Binding
		err      error
	}

	all := make(chan result)
	defer close(all)

	allApps := store.SortApps(p.store.App.AsMap())

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

	ret := []apps.Binding{}
	var problems error
	for i := 0; i < len(allApps); i++ {
		res := <-all
		ret = mergeBindings(ret, res.bindings)
		if res.err != nil {
			problems = multierror.Append(problems, errors.Wrap(res.err, string(res.appID)))
		}
	}
	return ret, problems
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}
