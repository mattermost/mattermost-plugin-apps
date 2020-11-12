package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	query := req.URL.Query()
	bindings, err := a.apps.API.GetBindings(&apps.Context{
		TeamID:       query.Get(apps.PropTeamID),
		ChannelID:    query.Get(apps.PropChannelID),
		ActingUserID: actingUserID,
		UserID:       actingUserID,
		PostID:       query.Get(apps.PropPostID),
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
