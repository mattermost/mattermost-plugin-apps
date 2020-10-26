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
	call := h.makeCall(PathMessage)
	modal := *call
	modal.AsModal = true

	httputils.WriteJSON(w,
		[]*api.Binding{
			{
				LocationID: api.LocationCommand,
				Bindings: []*api.Binding{
					{
						LocationID:  "message",
						Hint:        "[--user] message",
						Description: "send a message to a user",
						Call:        call,
					}, {
						LocationID:  "manage",
						Hint:        "subscribe | unsubscribe ",
						Description: "manage channel subscriptions to greet new users",
						Bindings: []*api.Binding{
							{
								LocationID:  "subscribe",
								Hint:        "[--channel]",
								Description: "subscribes a channel to greet new users",
								Call:        h.makeCall(PathMessage, "mode", "on"),
							}, {
								LocationID:  "unsubscribe",
								Hint:        "[--channel]",
								Description: "unsubscribes a channel from greeting new users",
								Call:        h.makeCall(PathMessage, "mode", "off"),
							},
						},
					},
				},
			}, {
				LocationID: api.LocationPostMenu,
				Bindings: []*api.Binding{
					{
						LocationID:  "message",
						Description: "message a user",
						Call:        &modal,
					},
				},
			},
		})
	return http.StatusOK, nil
}
