package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	query := req.URL.Query()
	bindings, err := a.apps.API.GetBindings(&api.Context{
		TeamID:       query.Get(constants.TeamID),
		ChannelID:    query.Get(constants.ChannelID),
		ActingUserID: actingUserID,
		UserID:       actingUserID,
		PostID:       query.Get(constants.PostID),
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
