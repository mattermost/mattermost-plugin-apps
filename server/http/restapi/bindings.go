package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	query := req.URL.Query()
	bindings, err := a.proxy.GetBindings(&apps.Context{
		ActingUserID:      actingUserID,
		ChannelID:         query.Get(api.PropChannelID),
		MattermostSiteURL: a.conf.GetConfig().MattermostSiteURL,
		PostID:            query.Get(api.PropPostID),
		TeamID:            query.Get(api.PropTeamID),
		UserAgent:         query.Get(api.PropUserAgent),
		UserID:            actingUserID,
	})
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
