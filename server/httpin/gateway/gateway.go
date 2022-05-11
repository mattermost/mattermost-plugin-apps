package gateway

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
)

type gateway struct {
	*handler.Handler
}

func Init(h *handler.Handler) {
	g := gateway{
		Handler: h.PathPrefix(path.Apps),
	}

	// Static
	h.HandleFunc("/{appid}/"+path.StaticFolder+"/{name}",
		g.static, h.RequireActingUser).Methods(http.MethodGet)

	// Incoming remote webhooks
	h.HandleFunc("/{appid}"+path.Webhook,
		g.handleWebhook).Methods(http.MethodPost)
	h.HandleFunc("/{appid}"+path.Webhook+"/{path}",
		g.handleWebhook).Methods(http.MethodPost)

	// Remote OAuth2
	h.HandleFunc("/{appid}"+path.RemoteOAuth2Connect,
		g.remoteOAuth2Connect, h.RequireActingUser).Methods(http.MethodGet)
	h.HandleFunc("/{appid}"+path.RemoteOAuth2Complete,
		g.remoteOAuth2Complete, h.RequireActingUser).Methods(http.MethodGet)
}

func appIDVar(req *http.Request) apps.AppID {
	s, ok := mux.Vars(req)["appid"]
	if ok {
		return apps.AppID(s)
	}
	return ""
}
