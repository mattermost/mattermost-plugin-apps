package main

import (
	"embed"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

const iconPath = "icon.png"

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

func main() {
	app := goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       "example-expand",
			Version:     "v1.1.0",
			DisplayName: "A Mattermost app illustrating how `Call.Expand` works",
			Icon:        "icon.png",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/expand",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
		goapp.WithStatic(static),
		goapp.WithCommand(
			userAction(),
			// notify(),
		),
		goapp.WithChannelHeader(userAction()),
		goapp.WithPostMenu(userAction()),
	)

	app.HandleCall("/echo", handleEcho)
	// app.HandleCall("/event", handleEvent)

	app.RunHTTP()
}
