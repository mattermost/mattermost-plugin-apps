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
	m := apps.Manifest{
		AppID:       "hello-goapp",
		Version:     "v1.0.0",
		DisplayName: "Hello, world! as a goapp",
		Icon:        "icon.png",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/goapp",
		RequestedPermissions: []apps.Permission{
			apps.PermissionActAsBot,
		},
	}

	app := goapp.NewApp(m, utils.MustMakeCommandLogger(zapcore.DebugLevel)).
		WithStatic(static).
		WithCommand(send).
		WithChannelHeader(send)

	panic(app.RunHTTP())
}

var send = goapp.NewBindableForm("send",
	handleSend,
	apps.Form{
		Title: "Hello, world!",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Type: "text",
				Name: "message",
			},
		},
		Submit: apps.NewCall("/send").WithExpand(apps.Expand{ActingUserAccessToken: apps.ExpandAll}),
	})

func handleSend(creq goapp.CallRequest) apps.CallResponse {
	message := "Hello, world!"
	custom := creq.GetValue("message", "")
	if custom != "" {
		message += " ...and " + custom + "!"
	}
	creq.AsBot().DM(creq.Context.ActingUser.Id, message)
	return apps.NewTextResponse("Created a post in your DM channel. Message: `%s`.", message)
}
