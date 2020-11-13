package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request) {
	call, err := apps.UnmarshalCallFromReader(req.Body)
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
	call.Context.ActingUserID = actingUserID

	res, err := a.apps.API.Call(call)
	if err != nil {
		httputils.WriteInternalServerError(w, err)
		return
	}

	httputils.WriteJSON(w, res)
}
