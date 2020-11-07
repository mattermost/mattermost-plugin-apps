package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	bindings, err := a.apps.API.GetBindings(&apps.Context{
		TeamID:       req.Form.Get(apps.PropTeamID),
		ChannelID:    req.Form.Get(apps.PropChannelID),
		ActingUserID: actingUserID,
		UserID:       actingUserID,
		PostID:       req.Form.Get(apps.PropPostID),
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
