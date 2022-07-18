package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func handleEcho(creq goapp.CallRequest) apps.CallResponse {
	originallCallJSON := utils.JSONBlock(creq.Call)
	contextJSON := utils.JSONBlock(creq.Context)

	return apps.NewTextResponse("Original Call:%s\n---\nResulting CallRequest.Context:%s",
		originallCallJSON, contextJSON)
}
