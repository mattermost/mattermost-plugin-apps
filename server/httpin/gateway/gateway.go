package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type gateway struct {
	proxy proxy.Service
}

func Init(c *incoming.Request, router *mux.Router, p proxy.Service, _ appservices.Service) {
	g := &gateway{
		proxy: p,
	}

	subrouter := router.PathPrefix(path.Apps).Subrouter()

	// Static
	subrouter.Handle("/{appid}/"+path.StaticFolder+"/{name}",
		incoming.AddContext(g.static, c).RequireUser()).Methods(http.MethodGet)

	// Incoming remote webhooks
	subrouter.Handle("/{appid}"+path.Webhook,
		incoming.AddContext(g.handleWebhook, c)).Methods(http.MethodPost)
	subrouter.Handle("/{appid}"+path.Webhook+"/{path}",
		incoming.AddContext(g.handleWebhook, c)).Methods(http.MethodPost)

	// Remote OAuth2
	subrouter.Handle("/{appid}"+path.RemoteOAuth2Connect,
		incoming.AddContext(g.remoteOAuth2Connect, c).RequireUser()).Methods(http.MethodGet)
	subrouter.Handle("/{appid}"+path.RemoteOAuth2Complete,
		incoming.AddContext(g.remoteOAuth2Complete, c).RequireUser()).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
