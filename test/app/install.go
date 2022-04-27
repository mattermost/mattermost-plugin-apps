package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func handleInstall(c *apps.CallRequest) apps.CallResponse {
	Log.Debugf("handleInstall called")
	return apps.NewTextResponse("handleInstall called")
}
