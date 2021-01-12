package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func Bindings() []*api.Binding {
	justSend := api.MakeCall(PathSendSurvey)

	modal := api.MakeCall(PathSendSurvey)
	modal.Type = api.CallTypeForm

	modalFromPost := api.MakeCall(PathSendSurvey)
	modalFromPost.Type = api.CallTypeForm
	modalFromPost.Expand = &api.Expand{Post: api.ExpandAll}
	return []*api.Binding{
		{
			// TODO make this a subscribe button, with a state (current subscription status)
			Location: api.LocationChannelHeader,
			Bindings: []*api.Binding{
				{
					Location:     "send",
					Label:        "Survey a user",
					Icon:         "https://www.clipartmax.com/png/middle/243-2431175_hello-hello-icon-png.png",
					Hint:         "Send survey to a user",
					Description:  "Send a customized emotional response survey to a user",
					Call:         modal,
					Presentation: "modal",
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
					Call:        justSend, // will use ActingUserID by default
				},
				{
					Location:     "send",
					Label:        "Survey a user",
					Hint:         "Send survey to a user",
					Description:  "Send a customized emotional response survey to a user",
					Call:         modalFromPost,
					Presentation: "modal",
				},
			},
		},
		// TODO /Command binding is a placeholder, may not be final, test!
		{
			Location: api.LocationCommand,
			Bindings: []*api.Binding{
				{
					Label:       "message",
					Location:    "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        justSend,
				}, {
					Label:       "manage",
					Location:    "manage",
					Hint:        "subscribe | unsubscribe ",
					Description: "manage channel subscriptions to greet new users",
					Bindings: []*api.Binding{
						{
							Label:       "subscribe",
							Location:    "subscribe",
							Hint:        "[--channel]",
							Description: "subscribes a channel to greet new users",
							Call:        api.MakeCall(PathSubscribeChannel, "mode", "on"),
						}, {
							Label:       "unsubscribe",
							Location:    "unsubscribe",
							Hint:        "[--channel]",
							Description: "unsubscribes a channel from greeting new users",
							Call:        api.MakeCall(PathSubscribeChannel, "mode", "off"),
						},
					},
				},
			},
		},
	}
}
