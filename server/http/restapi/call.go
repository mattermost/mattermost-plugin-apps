package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	res := a.proxy.Call(sessionID(req), actingID(req), call)
	if res.Type == "" {
		res.Type = apps.CallResponseTypeOK
	}

	httputils.WriteJSON(w, res)
}
