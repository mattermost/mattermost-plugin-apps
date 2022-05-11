package gateway

import (
	"net/http"

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

	// Static: /{appid}/static/...
	h.HandleFunc(handler.AppIDPath+path.StaticFolder+"/{name}", g.static,
		h.RequireActingUser,
		h.RequireToAppFromPath,
	).Methods(http.MethodGet)

	// Incoming remote webhooks: /{appid}/webhook/...
	h.HandleFunc(handler.AppIDPath+path.Webhook, g.handleWebhook,
		h.RequireToAppFromPath,
	).Methods(http.MethodPost)

	h.HandleFunc(handler.AppIDPath+path.Webhook+"/{path}", g.handleWebhook,
		h.RequireToAppFromPath,
	).Methods(http.MethodPost)

	// Remote OAuth2: /{appid}/oauth2/remote/connect and /{appid}/oauth2/remote/complete
	h.HandleFunc(handler.AppIDPath+path.RemoteOAuth2Connect, g.remoteOAuth2Connect,
		h.RequireActingUser,
		h.RequireToAppFromPath,
	).Methods(http.MethodGet)

	h.HandleFunc(handler.AppIDPath+path.RemoteOAuth2Complete, g.remoteOAuth2Complete,
		h.RequireActingUser,
		h.RequireToAppFromPath,
	).Methods(http.MethodGet)
}
