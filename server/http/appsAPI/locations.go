package appsAPI

import (
	"errors"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *api) handleLocations(w http.ResponseWriter, req *http.Request, actingUserID string) {
	userID := req.URL.Query().Get("userID")
	if userID == "" {
		httputils.WriteBadRequestError(w, errors.New("no user id"))
		return
	}

	if userID != actingUserID {
		httputils.WriteUnauthorizedError(w, errors.New("user id is not the same"))
		return
	}

	channelID := req.URL.Query().Get("channelID")
	if channelID == "" {
		httputils.WriteBadRequestError(w, errors.New("no channel id"))
		return
	}

	locations, err := a.apps.API.GetLocations(userID, channelID)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, locations)
}
