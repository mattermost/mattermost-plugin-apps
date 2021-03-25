package restapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID, token string) {
	q := req.URL.Query()
	cc := a.conf.GetConfig().SetContextDefaults(&apps.Context{
		ActingUserID: actingUserID,
		TeamID:       q.Get(config.PropTeamID),
		ChannelID:    q.Get(config.PropChannelID),
		PostID:       q.Get(config.PropPostID),
		UserAgent:    q.Get(config.PropUserAgent),
		UserID:       actingUserID,
	})

	bindings, err := a.proxy.GetBindings(apps.SessionToken(token), cc)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}

func (a *restapi) handleInvalidateCache(w http.ResponseWriter, req *http.Request, actingUserID string, token string) {
	vars := mux.Vars(req)

	appID := vars["app_id"]
	userID := vars["user_id"]
	channelID := vars["channel_id"]

	if appID == "" {
		httputils.WriteBadRequestError(w, errors.New("app_id not specified"))
		return
	}

	if err := a.proxy.InvalidateCache(apps.AppID(appID), userID, channelID); err != nil {
		httputils.WriteInternalServerError(w, err)
	}
}