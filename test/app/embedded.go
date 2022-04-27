package main

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

func embeddedCommandBinding(_ apps.Context) apps.Binding {
	return apps.Binding{
		Label: "embedded",
		Bindings: []apps.Binding{
			{
				Label: "create",
				Icon:  "icon.png",
				Submit: apps.NewCall(CreateEmbedded).WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					Channel:               apps.ExpandSummary,
				}),
			},
		},
	}
}

func initHTTPEmbedded(r *mux.Router) {
	handleCall(r, CreateEmbedded, handleCreateEmbedded)
}

func handleCreateEmbedded(creq *apps.CallRequest) apps.CallResponse {
	client := appclient.AsActingUser(creq.Context)
	p := &model.Post{
		ChannelId: creq.Context.Channel.Id,
	}

	p.AddProp(apps.PropAppBindings, []apps.Binding{
		{
			Location:    "embedded",
			AppID:       creq.Context.AppID,
			Description: "Please fill out this form so we can get it fixed  :hammer_and_wrench:",
			Bindings: []apps.Binding{
				{
					Location: "problem",
					Bindings: []apps.Binding{
						{
							Location: "hardware",
							Submit:   callOK,
							Label:    "Hardware Failure",
						},
						{
							Location: "software",
							Label:    "Software Error",
							Submit:   callOK,
						},
						{
							Location: "wrong",
							Label:    "Wrong Product",
							Submit:   callOK,
						},
					},
				},
				{
					Location: "provider",
					Bindings: []apps.Binding{
						{
							Location: "work",
							Label:    "Cell Phone",
							Submit:   callOK,
						},
					},
				},
				{
					Location: "button",
					Label:    "Submit",
					Submit:   callOK,
				},
			},
		},
	})

	_, err := client.CreatePost(p)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewTextResponse("")
}
