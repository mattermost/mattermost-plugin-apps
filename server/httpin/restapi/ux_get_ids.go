package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *restapi) initGetBotIDs(h *httpin.Handler) {
	h.HandleFunc(path.BotIDs,
		a.GetBotIDs, httpin.RequireUser).Methods(http.MethodGet)
}

func (a *restapi) initGetOAuthAppIDs(h *httpin.Handler) {
	h.HandleFunc(path.OAuthAppIDs,
		a.GetOAuthAppIDs, httpin.RequireUser).Methods(http.MethodGet)
}

// GetBotIDs returns the list of all Apps' bot user IDs.
//   Path: /api/v1/bot-ids
//   Method: GET
//   Input: none
//   Output: []string - the list of Bot user IDs for all installed Apps.
func (a *restapi) GetBotIDs(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	apps := a.proxy.GetInstalledApps(req)
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
func (a *restapi) GetOAuthAppIDs(req *incoming.Request, w http.ResponseWriter, r *http.Request) {
	apps := a.proxy.GetInstalledApps(req)
	ids := []string{}
	for _, app := range apps {
		if app.MattermostOAuth2 != nil {
			ids = append(ids, app.MattermostOAuth2.Id)
		}
	}
	b, _ := json.Marshal(ids)
	_, _ = w.Write(b)
}
