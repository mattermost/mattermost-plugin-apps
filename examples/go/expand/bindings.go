package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

func getBindings(creq goapp.CallRequest) apps.CallResponse {
	return apps.NewDataResponse([]apps.Binding{
		*creq.App.CommandBindings(action.Binding(creq)),
		*creq.App.PostMenuBindings(action.Binding(creq)),
		*creq.App.ChannelHeaderBindings(action.Binding(creq)),
	})
}
