package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

//go:embed icon.png
var IconData []byte

// Manifest declares the app's metadata. It must be provided for the app to be
// installable. In this example, the following permissions are requested:
//   - Create posts as a bot.
//   - Add icons to the channel header that will call back into your app when
//     clicked.
//   - Add a /-command with a callback.
var Manifest = apps.Manifest{
	// App ID must be unique across all Mattermost Apps.
	AppID: "hello-world",

	// App's release/version.
	Version: "v1.0.0",

	// A (long) display name for the app.
	DisplayName: "Hello, world!",

	// The icon for the app's bot account, same icon is also used for bindings
	// and forms.
	Icon: "icon.png",

	// HomepageURL is required for an app to be installable.
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/hello-world",

	// Need ActAsBot to post back to the user.
	RequestedPermissions: []apps.Permission{
		apps.PermissionActAsBot,
		apps.PermissionActAsUser,
	},

	// Add UI elements: a /-command, and a channel header button.
	RequestedLocations: []apps.Location{
		apps.LocationChannelHeader,
		apps.LocationCommand,
	},

	// Running the app as an HTTP service is the only deployment option
	// supported.
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: "http://localhost:4000",
		},
	},
}

// The details for the App UI bindings
var Bindings = []apps.Binding{
	{
		Location: apps.LocationChannelHeader,
		Bindings: []apps.Binding{
			{
				Location: "send-button",        // an app-chosen string.
				Icon:     "icon.png",           // reuse the App icon for the channel header.
				Label:    "send hello message", // appearance in the "more..." menu.
				Form:     &SendForm,            // the form to display.
			},
		},
	},
	{
		Location: "/command",
		Bindings: []apps.Binding{
			{
				// For commands, Location is not necessary, it will be defaulted to the label.
				Icon:        "icon.png",
				Label:       "helloworld",
				Description: "Hello World app", // appears in autocomplete.
				Hint:        "[send]",          // appears in autocomplete, usually indicates as to what comes after choosing the option.
				Bindings: []apps.Binding{
					{
						Label: "send", // "/helloworld send" sub-command.
						Form:  &SendForm,
					},
				},
			},
		},
	},
}

// SendForm is used to display the modal after clicking on the channel header
// button. It is also used for `/helloworld send` sub-command's autocomplete. It
// contains just one field, "message" for the user to customize the message.
var SendForm = apps.Form{
	Title: "Hello, world!",
	Icon:  "icon.png",
	Fields: []apps.Field{
		{
			Type: "text",
			Name: "message",
		},
	},
	Submit: apps.NewCall("/send").WithExpand(apps.Expand{ActingUserAccessToken: apps.ExpandAll}),
}

// main sets up the http server, with paths mapped for the static assets, the
// bindings callback, and the send function.
func main() {
	// Serve static assets: the manifest and the icon.
	http.HandleFunc("/manifest.json",
		httputils.DoHandleJSON(Manifest))
	http.HandleFunc("/static/icon.png",
		httputils.DoHandleData("image/png", IconData))

	// Bindinings callback.
	http.HandleFunc("/bindings",
		httputils.DoHandleJSON(apps.NewDataResponse(Bindings)))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send", Send)

	addr := ":4000" // matches manifest.json
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://localhost" + addr + "/manifest.json' to install the app") // matches manifest.json
	log.Fatal(http.ListenAndServe(addr, nil))
}

// Send sends a direct message (DM) back to the user.
func Send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	// Customize the message, as provided in creq.Values, or use the default.
	message := "Hello, world!"
	v := creq.GetValue("message", "")
	if v != "" {
		message += fmt.Sprintf(" ...and **%s**!", v)
	}

	// Send it as a direct message to the user, from the app's bot.
	appclient.AsBot(creq.Context).DM(creq.Context.ActingUser.Id, message)

	// Respond from the user back to the bot. 
	appclient.AsActingUser(creq.Context).DM(creq.Context.BotUserID, "Hello back at you, bot!")

	// Respond with an ephemeral message, in the current channel.
	httputils.WriteJSON(w, apps.NewTextResponse("Created a post in your DM channel."))
}
