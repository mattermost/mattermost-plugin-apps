package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	bindings, err := a.apps.API.GetBindings(&api.Context{
		TeamID:       req.Form.Get(constants.TeamID),
		ChannelID:    req.Form.Get(constants.ChannelID),
		ActingUserID: actingUserID,
		UserID:       actingUserID,
		PostID:       req.Form.Get(constants.PostID),
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
