package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

// IconData contains the bot's icon data
//
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
	Version: "v1.2.0",

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

// Bindings contain the details for the App UI components
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
	Submit: apps.NewCall("/send").WithExpand(apps.Expand{
		ActingUser:            apps.ExpandID.Required(),
		ActingUserAccessToken: apps.ExpandAll.Required(),
	}),
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
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://localhost" + addr + "/manifest.json' to install the app") // matches manifest.json
	panic(server.ListenAndServe())
}

// Send sends a DM back to the user.
func Send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	_ = json.NewDecoder(req.Body).Decode(&c)

	message := "Hello, world!"
	v, ok := c.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}
	_, err := appclient.AsBot(c.Context).DM(c.Context.ActingUser.Id, message)
	if err != nil {
		_ = httputils.WriteJSON(w, apps.NewErrorResponse(errors.Wrap(err, "Failed to send bot DM")))
		return
	}

	_, err = appclient.AsActingUser(c.Context).DM(c.Context.BotUserID, "Hello, bot!")
	if err != nil {
		_ = httputils.WriteJSON(w, apps.NewErrorResponse(errors.Wrap(err, "Failed to respond to bot")))
		return
	}

	_ = httputils.WriteJSON(w, apps.NewTextResponse("Created a post in your DM channel."))
}
