package httpin

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// Call handles a call request for an App.
//
//	Path: /api/v1/call
//	Method: POST
//	Input: CallRequest
//	Output: CallResponse
func (s *Service) Call(r *incoming.Request, w http.ResponseWriter, req *http.Request) {
	creq, err := apps.CallRequestFromJSONReader(req.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal Call request")
		r.Log.WithError(err).Infof("incoming call failed")
		httputils.WriteErrorIfNeeded(w, utils.NewInvalidError(err))
		return
	}
	if creq.Context.UserAgentContext.AppID == "" {
		err = errors.New("app ID is not set in Call request")
		r.Log.WithError(err).Infof("incoming call failed")
		httputils.WriteErrorIfNeeded(w, utils.NewInvalidError(err))
		return
	}
	r = r.WithDestination(creq.Context.UserAgentContext.AppID)

	// Call the app.
	cresp := s.Proxy.InvokeCall(r, *creq)

	// <>/<> TODO move to proxy.
	// Only track submit calls.
	if creq.Context.UserAgentContext.TrackAsSubmit {
		s.Config.Telemetry().TrackCall(string(creq.Context.AppID), string(creq.Context.Location), r.ActingUserID(), "submit")
	}

	_ = httputils.WriteJSON(w, cresp)
}
