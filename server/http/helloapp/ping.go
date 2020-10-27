package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (h *helloapp) handlePing(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, call *api.Call) (int, error) {
	userID := call.Values.Get("user_id")
	if userID == "" {
		userID = call.Context.ActingUserID
	}
	h.ping(userID)
	httputils.WriteJSON(w, api.CallResponse{
		Type: api.CallResponseTypeOK,
	})
	return http.StatusOK, nil
}

func (h *helloapp) ping(userID string) {
	h.DM(userID, "PING message")
}
