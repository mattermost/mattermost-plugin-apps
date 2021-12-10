package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *restapi) initGetBotIDs(api *mux.Router) {
	api.HandleFunc(path.BotIDs,
		proxy.RequireUser(a.GetBotIDs)).Methods("GET")
}

func (a *restapi) initGetOAuthAppIDs(api *mux.Router) {
	api.HandleFunc(path.OAuthAppIDs,
		proxy.RequireUser(a.GetOAuthAppIDs)).Methods("GET")
}

// GetBotIDs returns the list of all Apps' bot user IDs.
//   Path: /api/v1/bot-ids
//   Method: GET
//   Input: none
//   Output: []string - the list of Bot user IDs for all installed Apps.
func (a *restapi) GetBotIDs(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
	apps := a.proxy.GetInstalledApps()
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
func (a *restapi) GetOAuthAppIDs(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
	apps := a.proxy.GetInstalledApps()
	ids := []string{}
	for _, app := range apps {
		if app.MattermostOAuth2.ClientID != "" {
			ids = append(ids, app.MattermostOAuth2.ClientID)
		}
	}
	b, _ := json.Marshal(ids)
	_, _ = w.Write(b)
}
