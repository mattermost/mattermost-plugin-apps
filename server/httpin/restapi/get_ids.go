package restapi

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/proxy"
)

func (a *restapi) handleGetBotIDs(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
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

func (a *restapi) handleGetOAuthAppIDs(w http.ResponseWriter, r *http.Request, _ proxy.Incoming) {
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
