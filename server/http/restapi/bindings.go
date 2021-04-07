package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, token, actingUserID string) {
	q := req.URL.Query()
	cc := a.conf.GetConfig().SetContextDefaults(&apps.Context{
		ActingUserID: actingUserID,
		TeamID:       q.Get(config.PropTeamID),
		ChannelID:    q.Get(config.PropChannelID),
		PostID:       q.Get(config.PropPostID),
		UserAgent:    q.Get(config.PropUserAgent),
		UserID:       actingUserID,
	})

	bindings, err := a.proxy.GetBindings(sessionID(req), actingID(req), cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
