package restapi

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
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

	query := req.URL.Query()
	bindings, err := a.proxy.GetBindings(apps.SessionToken(session.Token),
		&apps.Context{
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
