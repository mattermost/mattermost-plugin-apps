package proxy

import (
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
func (p *Proxy) GetBindings(r *incoming.Request, cc apps.Context) ([]apps.Binding, error) {
	all := make(chan []apps.Binding)
	defer close(all)

	allApps := store.SortApps(p.store.App.AsMap())
	for i := range allApps {
		app := allApps[i]
		go func(app apps.App) {
			bb, err := p.InvokeGetBindings(r.WithDestination(app.AppID), cc)
			if err != nil {
				r.Log.WithError(err).Debugf("failed to fetch app bindings")
			}
			all <- bb
		}(app)
	}

	ret := []apps.Binding{}
	for i := 0; i < len(allApps); i++ {
		bb := <-all
		ret = mergeBindings(ret, bb)
	}
	return ret, nil
}

func (p *Proxy) dispatchRefreshBindingsEvent(userID string) {
	if userID != "" {
		p.conf.MattermostAPI().Frontend.PublishWebSocketEvent(
			config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})
	}
}
