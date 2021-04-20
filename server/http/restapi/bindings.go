package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	q := req.URL.Query()
	cc := a.conf.GetConfig().SetContextDefaults(&apps.Context{
		ActingUserID: actingUserID,
		TeamID:       q.Get(config.PropTeamID),
		ChannelID:    q.Get(config.PropChannelID),
		PostID:       q.Get(config.PropPostID),
		UserAgent:    q.Get(config.PropUserAgent),
		UserID:       actingUserID,
	})

	bindings, err := a.proxy.GetBindings(sessionID, actingID(req), cc)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}

func (a *restapi) handleRefreshBindings(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	userID := q.Get("user_id")

	if userID == "" {
		httputils.WriteError(w, errors.New("no user_id provided"))
		return
	}

	_, err := a.mm.User.Get(userID)
	if err != nil {
		httputils.WriteError(w, errors.Wrapf(err, "failed to get user with id %v", userID))
		return
	}

	a.mm.Frontend.PublishWebSocketEvent(config.WebSocketEventRefreshBindings, map[string]interface{}{}, &model.WebsocketBroadcast{UserId: userID})

	w.Header().Add("Content-Type", "application/json")
	httputils.WriteJSON(w, map[string]interface{}{
		"success": true,
	})
}
