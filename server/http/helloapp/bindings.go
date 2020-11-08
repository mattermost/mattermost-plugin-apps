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
	sendSurvey := h.makeCall(PathSendSurvey)

	c := *sendSurvey
	sendSurveyModal := &c
	sendSurveyModal.Type = api.CallTypeForm

	out := []*api.Binding{
		{
			// TODO make this a subscribe button, with a state (current subscription status)
			Location: api.LocationChannelHeader,
			Bindings: []*api.Binding{
				{
					Location:    "send",
					Label:       "Survey a user",
					Icon:        "https://raw.githubusercontent.com/mattermost/mattermost-plugin-jira/master/assets/icon.svg",
					Hint:        "Send survey to a user",
					Description: "Send a customized emotional response survey to a user",
					Call:        sendSurvey, // should be Modal eventually
				},
			},
		}, {
			Location: api.LocationPostMenu,
			Bindings: []*api.Binding{
				{
					Location:    "send-me",
					Label:       "Survey myself",
					Hint:        "Send survey to myself",
					Description: "Send a customized emotional response survey to myself",
					Call:        sendSurvey, // will use ActingUserID by default
				},
				{
					Location:    "send",
					Label:       "Survey a user",
					Hint:        "Send survey to a user",
					Description: "Send a customized emotional response survey to a user",
					Call:        sendSurveyModal,
				},
			},
		},
		// TODO /Command binding is a placeholder, may not be final, test!
		{
			Location: api.LocationCommand,
			Bindings: []*api.Binding{
				{
					Label:       "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        sendSurvey,
				}, {
					Location:    "manage",
					Hint:        "subscribe | unsubscribe ",
					Description: "manage channel subscriptions to greet new users",
					Bindings: []*api.Binding{
						{
							Label:       "subscribe",
							Hint:        "[--channel]",
							Description: "subscribes a channel to greet new users",
							Call:        h.makeCall(PathSubscribeChannel, "mode", "on"),
						}, {
							Label:       "unsubscribe",
							Hint:        "[--channel]",
							Description: "unsubscribes a channel from greeting new users",
							Call:        h.makeCall(PathSubscribeChannel, "mode", "off"),
						},
					},
				},
			},
		},
	}

	httputils.WriteJSON(w, out)
	return http.StatusOK, nil
}
