package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCall(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	cc, err := a.proxy.CleanUserCallContext(actingUserID, call.Context)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user")))
		return
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	call.Context = cc
	res := a.proxy.Call(sessionID, actingUserID, call)
	httputils.WriteJSON(w, res)
}
