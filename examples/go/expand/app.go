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
	simplifiedManifest := apps.Manifest{
		AppID:       "example-expand",
		Version:     "v1.0.0",
		DisplayName: "Example of hwow Expand works in Calls",
		Icon:        "icon.png",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/expand",
		RequestedPermissions: []apps.Permission{
			apps.PermissionActAsBot,
			apps.PermissionActAsUser,
		},
	}

	app := goapp.NewApp(simplifiedManifest).
		WithStatic(static).
		WithAppCommand("", userAction)
		// .
		// WithChannelHeaderButton(noExpand).
		// WithPostMenu(noExpand)

	app.HandleCall("/echo", handleEcho)
		
	panic(app.RunHTTP())
}
