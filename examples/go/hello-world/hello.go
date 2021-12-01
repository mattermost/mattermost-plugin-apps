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

//go:embed bindings.json
var bindingsData []byte

//go:embed send_form.json
var formData []byte

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.HandleJSONData(manifestData))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings", httputils.HandleJSONData(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", httputils.HandleJSONData(formData))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", send)

	// Forces the send form to be displayed as a modal.
	http.HandleFunc("/send-modal/submit", httputils.HandleJSONData(formData))

	// Serves the icon for the app.
	http.HandleFunc("/static/icon.png",
		httputils.HandleData("image/png", iconData))

	addr := ":4000" // matches manifest.json
	fmt.Println("Listening on", addr)
	fmt.Println("Use '/apps install http http://localhost" + addr + "/manifest.json' to install the app") // matches manifest.json
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

	appclient.AsActingUser(c.Context).DM(c.Context.BotUserID, "Hello, bot!")

	httputils.WriteJSON(w,
		apps.NewTextResponse("Created a post in your DM channel."))
}
