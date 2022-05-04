package main

import (
	"embed"

	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

// main starts the app, as a standalone HTTP server. Use $PORT and $ROOT_URL to
// customize.
func main() {
	// The app's minimal manifest.
	m := apps.Manifest{
		AppID:       "hello-goapp",
		Version:     "v1.0.0",
		DisplayName: "Hello, world! as a goapp",
		Icon:        "icon.png",
		HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/goapp",
	}

	// send is the bindable (form) action that implements the /hello-goapp send
	// command.
	send := goapp.NewBindableForm("send", handleSend, apps.Form{
		Title:  "Hello, world!",
		Icon:   "icon.png",
		Fields: []apps.Field{{Name: "message"}},
	})

	// Create the app, add `send` to the app's command and the channel header.
	app := goapp.NewApp(m, utils.MustMakeCommandLogger(zapcore.DebugLevel)).
		WithStatic(static).
		WithCommand(send).
		WithChannelHeader(send)

	// Run the app.
	panic(app.RunHTTP())
}

// handleSend processes the send call.
func handleSend(creq goapp.CallRequest) apps.CallResponse {
	message := "Hello from a goapp."
	custom := creq.GetValue("message", "")
	if custom != "" {
		message += " ...and " + custom + "!"
	}
	creq.AsBot().DM(creq.Context.ActingUser.Id, message)
	return apps.NewTextResponse("Created a post in your DM channel. Message: `%s`.", message)
}
