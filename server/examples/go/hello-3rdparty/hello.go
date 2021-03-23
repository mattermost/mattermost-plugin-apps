package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
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
	// Static handlers

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	// Serve the form for sending a Hello message.
	http.HandleFunc("/send/form", writeJSON(formData))

	// Submit handlers

	// `send` command - send a Hello message.
	http.HandleFunc("/send/submit", send)

	// `connect` command - display the OAuth2 connect link.
	http.HandleFunc("/connect/submit", connect)

	// Handle an OAuth2 connect request redirect.
	http.HandleFunc("/oauth2/connect/submit", oauth2Success)

	// Handle a successful OAuth2 connection.
	http.HandleFunc("/oauth2/success/submit", oauth2Success)

	http.ListenAndServe(":8080", nil)
}

func send(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	message := "Hello, world!"
	v, ok := c.Values["message"]
	if ok && v != nil {
		message += fmt.Sprintf(" ...and %s!", v)
	}
	mmclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)

	json.NewEncoder(w).Encode(apps.CallResponse{})
}

func connect(w http.ResponseWriter, req *http.Request) {
	fmt.Println("<>/<> connect")
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	mmclient.AsBot(c.Context).DM(c.Context.ActingUserID,
		fmt.Sprintf("[Connect] your Google Calendar.](%s%s)", c.Context.MattermostSiteURL, path.Join(c.Context.AppPath, "/oauth2/connect")))

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: "connect link `TODO` markdown",
	})
}

func oauth2Success(w http.ResponseWriter, req *http.Request) {
	fmt.Println("<>/<> OAuth2 success")
	json.NewEncoder(w).Encode(apps.CallResponse{})
}

func writeData(ct string, data []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		w.Write(data)
	}
}

func writeJSON(data []byte) func(w http.ResponseWriter, r *http.Request) {
	return writeData("application/json", data)
}
