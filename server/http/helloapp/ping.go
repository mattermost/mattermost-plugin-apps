package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (h *helloapp) handlePing(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, data *apps.CallRequest) (int, error) {
	h.ping(data.Values.Get("userID"))
	return http.StatusOK, nil
}

func (h *helloapp) ping(userID string) {
	h.DM(userID, "PING message")
}
