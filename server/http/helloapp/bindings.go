package helloapp

import (
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

// Install function metadata is not necessary, but fillint it out (minimally)
// for demo purposes. Install does not bind to any locations, it's Expand is
// pre-determined by the server.
func (h *helloapp) bindings(w http.ResponseWriter, req *http.Request, claims *apps.JWTClaims, cc *apps.Context) (int, error) {
	sendSurvey := h.makeCall(PathSendSurvey)

	c := *sendSurvey
	c.Expand = &apps.Expand{Post: apps.ExpandAll}

	sendSurveyModal := &c
	sendSurveyModal.Type = apps.CallTypeForm

	out := []*apps.Binding{
		{
			// TODO make this a subscribe button, with a state (current subscription status)
			Location: apps.LocationChannelHeader,
			Bindings: []*apps.Binding{
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
			Location: apps.LocationPostMenu,
			Bindings: []*apps.Binding{
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
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					Label:       "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        sendSurvey,
				}, {
					Location:    "manage",
					Hint:        "subscribe | unsubscribe ",
					Description: "manage channel subscriptions to greet new users",
					Bindings: []*apps.Binding{
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
