package main

import (
	"embed"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-server/v6/model"
)

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

// main starts the app, as a standalone HTTP server. Use $PORT and $ROOT_URL to
// customize.
func main() {
	// Create the app, add `send` to the app's command and the channel header.
	goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       "hello-goapp",
			Version:     "v1.1.0",
			DisplayName: "Hello, world! as a goapp",
			Icon:        "icon.png",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/goapp",
			RequestedScopes: model.AppScopes{
				model.ScopeUsersRead,
				model.ScopePostsCreate,
				model.ScopeChannelsCreate,
			},
		},
		goapp.WithStatic(static),
		goapp.WithCommand(send),
		goapp.WithChannelHeader(send),
	).RunHTTP()
}

// send is the bindable (form) action that implements the /hello-goapp send
// command.
var send = goapp.MakeBindableFormOrPanic("send",
	apps.Form{
		Title:  "Hello, world!",
		Icon:   "icon.png",
		Fields: []apps.Field{{Name: "message"}},
		Submit: &apps.Call{
			Expand: &apps.Expand{
				ActingUser: apps.ExpandID.Required(),
			},
		},
	},
	func(creq goapp.CallRequest) apps.CallResponse {
		message := "Hello from a goapp."
		custom := creq.GetValue("message", "")
		if custom != "" {
			message += " ...and " + custom + "!"
		}

		creq.AsBot().DM(creq.Context.ActingUser.Id, message)
		return apps.NewTextResponse("Created a post in your DM channel. Message: `%s`.", message)
	},
)
