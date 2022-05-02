package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

func getBindings(creq goapp.CallRequest) apps.CallResponse {
	return apps.NewDataResponse([]apps.Binding{
		*creq.App.CommandBindings(noExpand.Binding(creq)),
		*creq.App.PostMenuBindings(noExpand.Binding(creq)),
		*creq.App.ChannelHeaderBindings(noExpand.Binding(creq)),
	})
}
