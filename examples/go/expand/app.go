package main

import (
	"embed"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"go.uber.org/zap/zapcore"
)

const iconPath = "icon.png"

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

func main() {
	simplifiedManifest := apps.Manifest{
		AppID:       "example-expand",
		Version:     "v1.0.0",
		DisplayName: "A Mattermost app illustrating how `Call.Expand` works",
		Icon:        "icon.png",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/expand",
		RequestedPermissions: []apps.Permission{
			apps.PermissionActAsBot,
			apps.PermissionActAsUser,
		},
	}

	app := goapp.NewApp(simplifiedManifest, utils.MustMakeCommandLogger(zapcore.DebugLevel)).
		WithStatic(static).
		WithCommand(userAction).
		WithChannelHeader(userAction).
		WithPostMenu(userAction)

	app.HandleCall("/echo", handleEcho)

	panic(app.RunHTTP())
}
