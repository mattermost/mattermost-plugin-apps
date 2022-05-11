package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *restapi) initGetBotIDs(h *handler.Handler) {
	h.HandleFunc(path.BotIDs,
		a.GetBotIDs, h.RequireActingUser).Methods(http.MethodGet)
}

func (a *restapi) initGetOAuthAppIDs(h *handler.Handler) {
	h.HandleFunc(path.OAuthAppIDs,
		a.GetOAuthAppIDs, h.RequireActingUser).Methods(http.MethodGet)
}

// GetBotIDs returns the list of all Apps' bot user IDs.
//   Path: /api/v1/bot-ids
//   Method: GET
//   Input: none
//   Output: []string - the list of Bot user IDs for all installed Apps.
func (a *restapi) GetBotIDs(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	apps, _ := a.Proxy.GetInstalledApps(r, false)
	ids := []string{}
	for _, app := range apps {
		if app.BotUserID != "" {
			ids = append(ids, app.BotUserID)
		}
	}
	b, _ := json.Marshal(ids)
	_, _ = w.Write(b)
}

// GetBindings returns combined bindings for all Apps.
//   Path: /api/v1/get-oauth-app-ids
//   Method: GET
//   Input: none
//   Output: []string - the list of OAuth ClientIDs for all installed Apps.
func (a *restapi) GetOAuthAppIDs(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	apps, _ := a.Proxy.GetInstalledApps(r, false)
	ids := []string{}
	for _, app := range apps {
		if app.MattermostOAuth2 != nil {
			ids = append(ids, app.MattermostOAuth2.Id)
		}
	}
	b, _ := json.Marshal(ids)
	_, _ = w.Write(b)
}
