package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func Bindings() []*api.Binding {
	sendSurvey := api.MakeCall(PathSendSurvey)

	clone := *sendSurvey
	clone.Expand = &api.Expand{Post: api.ExpandAll}

	sendSurveyModal := &clone
	sendSurveyModal.Type = api.CallTypeForm

	return []*api.Binding{
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
					Call:        sendSurveyModal,
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
					Location:    "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        sendSurvey,
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
