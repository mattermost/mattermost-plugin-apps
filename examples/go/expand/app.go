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

var Manifest = apps.Manifest{
	AppID:       "example-expand",
	Version:     "v1.0.0",
	DisplayName: "Example of hwow Expand works in Calls",
	Icon:        "icon.png",
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/expand",
	RequestedPermissions: []apps.Permission{
		apps.PermissionActAsBot,
		apps.PermissionActAsUser,
	},
	RequestedLocations: []apps.Location{
		apps.LocationChannelHeader,
		apps.LocationPostMenu,
		apps.LocationCommand,
	},
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: "http://localhost:8085",
		},
	},
}

func main() {
	app := goapp.NewApp(Manifest).
		WithStatic(static).
		WithIcon(iconPath)

	// Bindings.
	app.HandleCall("/bindings", getBindings)

	app.Handle(noExpand)

	panic(app.RunHTTP())
}
