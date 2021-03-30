package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		err = errors.Wrap(err, "Failed to unmarshal Call struct")
		httputils.WriteBadRequestError(w, err)
		return
	}

	actingUserID := req.Header.Get("Mattermost-User-Id")
	if actingUserID == "" {
		err = errors.New("user not logged in")
		httputils.WriteUnauthorizedError(w, err)
		return
	}

	if call.Context == nil || call.Context.AppID == "" {
		err = errors.New("must provide Context and set the app ID")
		httputils.WriteBadRequestError(w, err)
		return
	}
	call.Context.ActingUserID = actingUserID

	sessionID := req.Header.Get("MM_SESSION_ID")
	if sessionID == "" {
		err = errors.New("no user session")
		httputils.WriteUnauthorizedError(w, err)
		return
	}
	session, err := a.mm.Session.Get(sessionID)
	if err != nil {
		httputils.WriteUnauthorizedError(w, err)
		return
	}

	res := a.proxy.Call(apps.SessionToken(session.Token), call)
	if res.Type == "" {
		res.Type = apps.CallResponseTypeOK
	}

	httputils.WriteJSON(w, res)
}
