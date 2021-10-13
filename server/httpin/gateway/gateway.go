package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type gateway struct {
	conf  config.Service
	proxy proxy.Service
}

func Init(router *mux.Router, conf config.Service, p proxy.Service, _ appservices.Service) {
	g := &gateway{
		conf:  conf,
		proxy: p,
	}

	subrouter := router.PathPrefix(path.Apps).Subrouter()

	// Static
	subrouter.HandleFunc("/{appid}/"+path.StaticFolder+"/{name}",
		proxy.RequireUser(g.static)).Methods(http.MethodGet)

	// Incoming remote webhooks
	subrouter.HandleFunc("/{appid}"+path.Webhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	subrouter.HandleFunc("/{appid}"+path.RemoteOAuth2Connect,
		proxy.RequireUser(g.remoteOAuth2Connect)).Methods(http.MethodGet)
	subrouter.HandleFunc("/{appid}"+path.RemoteOAuth2Complete,
		proxy.RequireUser(g.remoteOAuth2Complete)).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
