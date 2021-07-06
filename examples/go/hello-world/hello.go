package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

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

const (
	host = "localhost"
	port = 8080
)

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", writeJSON(formData))

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", send)

	// Forces the send form to be displayed as a modal.
	http.HandleFunc("/send-modal/submit", writeJSON(formData))

	// Serves the icon for the app.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	addr := fmt.Sprintf("%v:%v", host, port)
	fmt.Printf(`hello-world app listening at http://%s`, addr)
	http.ListenAndServe(addr, nil)
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

	json.NewEncoder(w).Encode(apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: "Created a post in your DM channel.",
	})
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
