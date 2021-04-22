package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func (a *restapi) handleCallHTTP(w http.ResponseWriter, req *http.Request, sessionID, actingUserID string) {
	call, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		httputils.WriteError(w, utils.NewInvalidError(errors.Wrap(err, "failed to unmarshal Call request")))
		return
	}

	res, err := a.handleCall(sessionID, actingUserID, call)
	if err != nil {
		httputils.WriteError(w, err)
		return
	}

	httputils.WriteJSON(w, res)
}

func (a *restapi) handleCall(sessionID, actingUserID string, call *apps.CallRequest) (*apps.ProxyCallResponse, error) {
	cc, err := a.proxy.CleanUserCallContext(actingUserID, call.Context)
	if err != nil {
		return nil, utils.NewInvalidError(errors.Wrap(err, "invalid call context for user"))
	}

	cc = a.conf.GetConfig().SetContextDefaults(cc)

	call.Context = cc
	res := a.proxy.Call(sessionID, actingUserID, call)
	return res, nil
}
