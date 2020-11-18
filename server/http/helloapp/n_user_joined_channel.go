package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
)

func (h *helloapp) nUserJoinedChannel(w http.ResponseWriter, req *http.Request,
	claims *apps.JWTClaims, n *apps.Notification) (int, error) {
	go func() {
		err := h.sendSurvey(n.Context.UserID, "welcome to channel")
		if err != nil {
			h.apps.Mattermost.Log.Error("error sending survey", "err", err.Error())
		}
	}()
	return http.StatusOK, nil
}
