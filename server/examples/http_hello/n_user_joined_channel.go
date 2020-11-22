package http_hello

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func (h *helloapp) nUserJoinedChannel(w http.ResponseWriter, req *http.Request,
	claims *api.JWTClaims, n *api.Notification) (int, error) {
	go func() {
		err := h.sendSurvey(n.Context.UserID, "welcome to channel")
		if err != nil {
			h.api.Mattermost.Log.Error("error sending survey", "err", err.Error())
		}
	}()
	return http.StatusOK, nil
}
