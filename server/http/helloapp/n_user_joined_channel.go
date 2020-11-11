package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (h *helloapp) nUserJoinedChannel(w http.ResponseWriter, req *http.Request,
	claims *apps.JWTClaims, n *apps.Notification) (int, error) {
	go h.sendSurvey(n.Context.UserID, "welcome to channel")
	return http.StatusOK, nil
}
