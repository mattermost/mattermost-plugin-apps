package restapi

import (
	"encoding/json"
	"net/http"
)

func (a *restapi) handleGetBotIDs(w http.ResponseWriter, r *http.Request) {
	apps := a.proxy.GetInstalledApps()
	ids := []string{}
	for _, app := range apps {
		if app.BotUserID != "" {
			ids = append(ids, app.BotUserID)
		}
	}
	b, _ := json.Marshal(ids)
	w.Write(b)
}

func (a *restapi) handleGetOAuthAppIDs(w http.ResponseWriter, r *http.Request) {
	apps := a.proxy.GetInstalledApps()
	ids := []string{}
	for _, app := range apps {
		if app.MattermostOAuth2.ClientID != "" {
			ids = append(ids, app.MattermostOAuth2.ClientID)
		}
	}
	b, _ := json.Marshal(ids)
	w.Write(b)
}
