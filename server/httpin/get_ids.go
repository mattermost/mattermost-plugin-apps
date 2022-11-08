package httpin

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

// GetBotIDs returns the list of all Apps' bot user IDs.
//
//	Path: /api/v1/bot-ids
//	Method: GET
//	Input: none
//	Output: []string - the list of Bot user IDs for all installed Apps.
func (s *Service) GetBotIDs(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	apps := s.Proxy.GetInstalledApps()
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
//
//	Path: /api/v1/get-oauth-app-ids
//	Method: GET
//	Input: none
//	Output: []string - the list of OAuth ClientIDs for all installed Apps.
func (s *Service) GetOAuthAppIDs(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	apps := s.Proxy.GetInstalledApps()
	ids := []string{}
	for _, app := range apps {
		if app.MattermostOAuth2 != nil {
			ids = append(ids, app.MattermostOAuth2.Id)
		}
	}
	b, _ := json.Marshal(ids)
	_, _ = w.Write(b)
}
