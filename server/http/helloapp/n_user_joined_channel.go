package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (h *helloapp) handleUserJoinedChannel(w http.ResponseWriter, req *http.Request,
	claims *apps.JWTClaims, n *api.Notification) (int, error) {
	go h.message(n.Context.UserID, "welcome to channel")
	return http.StatusOK, nil
}
