package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type gateway struct {
	proxy proxy.Service
}

func Init(rh httpin.Handler, p proxy.Service, _ appservices.Service) {
	g := &gateway{
		proxy: p,
	}

	rh.Router = rh.Router.PathPrefix(path.Apps).Subrouter()

	// Static
	rh.HandleFunc("/{appid}/"+path.StaticFolder+"/{name}",
		g.static, httpin.RequireUser).Methods(http.MethodGet)

	// Incoming remote webhooks
	rh.HandleFunc("/{appid}"+path.Webhook,
		g.handleWebhook).Methods(http.MethodPost)
	rh.HandleFunc("/{appid}"+path.Webhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	rh.HandleFunc("/{appid}"+path.RemoteOAuth2Connect,
		g.remoteOAuth2Connect, httpin.RequireUser).Methods(http.MethodGet)
	rh.HandleFunc("/{appid}"+path.RemoteOAuth2Complete,
		g.remoteOAuth2Complete, httpin.RequireUser).Methods(http.MethodGet)
}

func appIDVar(r *http.Request) apps.AppID {
	s, ok := mux.Vars(r)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
