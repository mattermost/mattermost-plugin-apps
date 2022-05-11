package restapi

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/httpin/handler"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

func (a *restapi) initCall(h *handler.Handler) {
	h.HandleFunc(path.Call,
		a.Call, h.RequireActingUser).Methods(http.MethodPost)
}

// Call handles a call request for an App.
//   Path: /api/v1/call
//   Method: POST
//   Input: CallRequest
//   Output: CallResponse
func (a *restapi) Call(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	creq, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal Call request")
		r.Log.WithError(err).Infof("incoming call failed")
		httputils.WriteError(w, utils.NewInvalidError(err))
		return
	}

	// Call the app.
	cresp := a.Proxy.Call(r, creq.Context.UserAgentContext.AppID, *creq)

	// Only track submit calls.
	if creq.Context.UserAgentContext.TrackAsSubmit {
		a.Config.Telemetry().TrackCall(string(creq.Context.AppID), string(creq.Context.Location), r.ActingUserID(), "submit")
	}

	_ = httputils.WriteJSON(w, cresp)
}
