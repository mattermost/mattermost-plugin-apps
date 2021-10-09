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
var iconData []byte

//go:embed manifest.json
var manifestData []byte

var sendForm = apps.Form{
	Title: "Hello, world!",
	Icon:  "icon.png",
	Fields: []apps.Field{
		{
			Type:  "text",
			Name:  "message",
			Label: "message",
		},
	},
	Submit: &apps.Call{
		Path: "/send",
	},
}

var bindings = []apps.Binding{
	{
		Location: "/channel_header",
		Bindings: []apps.Binding{
			{
				Location: "send-button",
				Icon:     "icon.png",
				Label:    "send hello message",
				Form:     &sendForm,
			},
		},
	},
	{
		Location: "/command",
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "helloworld",
				Description: "Hello World app",
				Hint:        "[send]",
				Bindings: []apps.Binding{
					{
						Location: "send",
						Label:    "send",
						Form:     &sendForm,
					},
				},
			},
		},
	},
}

func main() {
	// Serve static assets: the manifest and the icon.
	http.HandleFunc("/manifest.json",
		httputils.HandleStaticJSONData(manifestData))
	http.HandleFunc("/static/icon.png",
		httputils.HandleStaticData("image/png", iconData))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings",
		httputils.HandleStaticJSON(bindings))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send", send)

	addr := ":4000" // matches manifest.json
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install url http://localhost" + addr + "/manifest.json' to install the app") // matches manifest.json
	log.Fatal(http.ListenAndServe(addr, nil))
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	message := "Hello, world!"
	v, ok := c.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}
	appclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}
