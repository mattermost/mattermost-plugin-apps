package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"

	"github.com/pkg/errors"
)

func (a *restapi) handleGetBindings(w http.ResponseWriter, req *http.Request, actingUserID string) {
	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		err := errors.New("no user session")
		httputils.WriteUnauthorizedError(w, err)
		return
	}
	session, err := a.mm.Session.Get(sessionID)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return
	}

	cc := a.conf.GetConfig().NewContext()
	q := req.URL.Query()
	cc.ActingUserID = actingUserID
	cc.TeamID = q.Get(config.PropTeamID)
	cc.ChannelID = q.Get(config.PropChannelID)
	cc.PostID = q.Get(config.PropPostID)
	cc.UserAgent = q.Get(config.PropUserAgent)
	cc.UserID = actingUserID

	bindings, err := a.proxy.GetBindings(apps.SessionToken(session.Token), cc)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, bindings)
}
