package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func handleBindings(creq *apps.CallRequest) apps.CallResponse {
	return apps.NewDataResponse([]apps.Binding{
		{
			Location: apps.LocationChannelHeader,
			Bindings: channelHeaderBindings(creq.Context),
		},
		{
			Location: apps.LocationPostMenu,
			Bindings: postMenuBindings(creq.Context),
		},
		{
			Location: apps.LocationCommand,
			Bindings: commandBindings(creq.Context),
		},
	})
}

func newBinding(label, submitPath string) apps.Binding {
	return apps.Binding{
		Label:  label,
		Icon:   "icon.png",
		Submit: apps.NewCall(submitPath),
	}
}
