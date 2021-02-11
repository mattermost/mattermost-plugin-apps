package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	query := req.URL.Query()
	bindings, err := a.api.Proxy.GetBindings(&apps.Context{
		TeamID:            query.Get(api.PropTeamID),
		ChannelID:         query.Get(api.PropChannelID),
		ActingUserID:      actingUserID,
		UserID:            actingUserID,
		PostID:            query.Get(api.PropPostID),
		MattermostSiteURL: a.api.Configurator.GetConfig().MattermostSiteURL,
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
