package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

type gateway struct {
	conf  config.Service
	proxy proxy.Service
}

func Init(h *httpin.Handler, config config.Service, p proxy.Service, _ appservices.Service) {
	g := &gateway{
		conf:  config,
		proxy: p,
	}

	h = h.PathPrefix(path.Apps)

	// Static
	h.HandleFunc("/{appid}/"+path.StaticFolder+"/{name}",
		g.static, httpin.RequireUser).Methods(http.MethodGet)

	// Incoming remote webhooks
	h.HandleFunc("/{appid}"+path.Webhook,
		g.handleWebhook).Methods(http.MethodPost)
	h.HandleFunc("/{appid}"+path.Webhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	h.HandleFunc("/{appid}"+path.RemoteOAuth2Connect,
		g.remoteOAuth2Connect, httpin.RequireUser).Methods(http.MethodGet)
	h.HandleFunc("/{appid}"+path.RemoteOAuth2Complete,
		g.remoteOAuth2Complete, httpin.RequireUser).Methods(http.MethodGet)
}

func appIDVar(req *http.Request) apps.AppID {
	s, ok := mux.Vars(req)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
