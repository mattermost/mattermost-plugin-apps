package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

// Install function metadata is not necessary, but fillint it out (minimally)
// for demo purposes. Install does not bind to any locations, it's Expand is
// pre-determined by the server.
func (h *helloapp) handleBindings(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, cc *api.Context) (int, error) {
	sendSurvey := h.makeCall(pathSendSurvey)

	out := []*api.Binding{
		{
			// TODO make this a subscribe button, with a state (current subscription status)
			LocationID: api.LocationChannelHeader,
			Bindings: []*api.Binding{
				{
					LocationID:  "send",
					Description: "say hello to a user",
					Call:        sendSurvey,
					AsModal:     true,
				},
			},
		}, {
			LocationID: api.LocationPostMenu,
			Bindings: []*api.Binding{
				{
					LocationID:  "sendSurvey-me",
					Description: "say hello to myself",
					Call:        sendSurvey,
				},
				{
					LocationID:  "send",
					Description: "say hello to a user",
					Call:        sendSurvey,
				},
			},
		}, {
			LocationID: api.LocationCommand,
			Bindings: []*api.Binding{
				{
					LocationID:  "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        sendSurvey,
				}, {
					LocationID:  "manage",
					Hint:        "subscribe | unsubscribe ",
					Description: "manage channel subscriptions to greet new users",
					Bindings: []*api.Binding{
						{
							LocationID:  "subscribe",
							Hint:        "[--channel]",
							Description: "subscribes a channel to greet new users",
							Call:        h.makeCall(pathSubscribeChannel, "mode", "on"),
						}, {
							LocationID:  "unsubscribe",
							Hint:        "[--channel]",
							Description: "unsubscribes a channel from greeting new users",
							Call:        h.makeCall(pathSubscribeChannel, "mode", "off"),
						},
					},
				},
			},
		},
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}
