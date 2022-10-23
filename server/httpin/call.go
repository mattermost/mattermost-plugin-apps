package httpin

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// CallResponse contains everything the CallResponse struct contains, plus some additional
// data for the client, such as information about the App's bot account.
//
// Apps will use the CallResponse struct to respond to a CallRequest, and the proxy will
// decorate the response using the CallResponse to provide additional information.
type CallResponse struct {
	apps.CallResponse

	// Used to provide info about the App to client, e.g. the bot user id
	AppMetadata AppMetadataForClient `json:"app_metadata"`
}

type AppMetadataForClient struct {
	BotUserID   string `json:"bot_user_id,omitempty"`
	BotUsername string `json:"bot_username,omitempty"`
}

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
	app, cresp := s.Proxy.InvokeCall(r, *creq)
	ccresp := CallResponse{
		CallResponse: cresp,
	}
	if app != nil {
		ccresp.AppMetadata = AppMetadataForClient{
			BotUserID:   app.BotUserID,
			BotUsername: app.BotUsername,
		}
	}

	// <>/<> TODO move to proxy.
	// Only track submit calls.
	if creq.Context.UserAgentContext.TrackAsSubmit {
		s.Config.Telemetry().TrackCall(string(creq.Context.AppID), string(creq.Context.Location), r.ActingUserID(), "submit")
	}

	_ = httputils.WriteJSON(w, ccresp)
}
