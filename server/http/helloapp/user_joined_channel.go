package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (h *helloapp) handleUserJoinedChannel(w http.ResponseWriter, req *http.Request,
	claims *apps.JWTClaims, data *api.Notification) (int, error) {
	go h.ping(data.Context.UserID)
	return http.StatusOK, nil
}
