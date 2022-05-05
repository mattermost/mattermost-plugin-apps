package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// notify is the "container" for all /notify subcommands.
var notify = goapp.NewBindableMulti("notify", notifyUserHasBeenCreated).
	WithDescription("Example of how Expand works in subscriptions/notifications", "[ subcommand ]")

// notifyUserHasBeenCreated defines /notify user-has-been-created command.
var notifyUserHasBeenCreated = goapp.NewBindableAction("user-has-been-created", handleNotifyUserHasBeenCreated, nil).
	WithExpand(apps.Expand{
		ActingUser:            apps.ExpandSummary,
		ActingUserAccessToken: apps.ExpandAll,
	})

// handleUserAction processes the submit
func handleNotifyUserHasBeenCreated(creq goapp.CallRequest) apps.CallResponse {
	return apps.NewTextResponse("OK")
}
