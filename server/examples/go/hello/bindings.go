package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func Bindings() []*apps.Binding {
	justSend := apps.MakeCall(PathSendSurvey)

	modal := apps.MakeCall(PathSendSurveyModal)

	modalFromPost := apps.MakeCall(PathSendSurveyModal)
	modalFromPost.Expand = &apps.Expand{Post: apps.ExpandAll}

	commandToModal := apps.MakeCall(PathSendSurveyCommandToModal)
	return []*apps.Binding{
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
					Call:        modal,
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
					Call:        justSend, // will use ActingUserID by default
				},
				{
					Location:    "send",
					Label:       "Survey a user",
					Hint:        "Send survey to a user",
					Description: "Send a customized emotional response survey to a user",
					Call:        modalFromPost,
				},
			},
		},
		// TODO /Command binding is a placeholder, may not be final, test!
		{
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					Label:       "message",
					Location:    "message",
					Hint:        "[--user] message",
					Description: "send a message to a user",
					Call:        justSend,
				}, {
					Label:       "message-modal",
					Location:    "message-modal",
					Hint:        "[--message] message",
					Description: "send a message to a user",
					Call:        commandToModal,
				}, {
					Label:       "manage",
					Location:    "manage",
					Hint:        "subscribe | unsubscribe ",
					Description: "manage channel subscriptions to greet new users",
					Bindings: []*apps.Binding{
						{
							Label:       "subscribe",
							Location:    "subscribe",
							Hint:        "[--channel]",
							Description: "subscribes a channel to greet new users",
							Call:        apps.MakeCall(PathSubscribeChannel),
						}, {
							Label:       "unsubscribe",
							Location:    "unsubscribe",
							Hint:        "[--channel]",
							Description: "unsubscribes a channel from greeting new users",
							Call:        apps.MakeCall(PathUnsubscribeChannel),
						},
					},
				},
			},
		},
	}
}
