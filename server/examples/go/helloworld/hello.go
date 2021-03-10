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

//go:embed send_form.json
var formData []byte

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", manifest)

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", bindings)

	// The main form for sending a Hello message.
	http.HandleFunc("/send", send)

	// Forces the send form to be displayed as a modal.
	// TODO: ticket: this should be unnecessary.
	http.HandleFunc("/send-modal", sendModal)

	// Serves the icon for the App.
	http.HandleFunc("/static/icon.png", icon)

	http.ListenAndServe(":8080", nil)
}

func bindings(w http.ResponseWriter, req *http.Request) {
	bindings := []*apps.Binding{
		{
			Location: "/channel_header",
			// Make this a top-level command, not subcommand
			Bindings: []*apps.Binding{
				{
					//TODO: ticket: Why is Location necessary? (getting Not Found without it)
					//TODO: ticket: Location on call was not FQ
					Location: "send-button",
					//TODO: ticket: relative URL doesn't work
					Icon: "http://localhost:8080/static/icon.png",
					Call: &apps.Call{
						Path: "/send-modal",
					},
				},
			},
		},
		{
			Location: "/command",
			Bindings: []*apps.Binding{
				{
					//TODO: ticket: Location on call was "/command", not "/command/helloworld/send"
					Location: "send",
					//TODO: ticket: Location alone does not work, requires Label
					Label: "send",
					Call: &apps.Call{
						Path: "/send",
					},
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		Type: "ok",
		Data: bindings,
	})
}

func send(w http.ResponseWriter, req *http.Request) {
	call := apps.Call{}
	_ = json.NewDecoder(req.Body).Decode(&call)

	w.Header().Set("Content-Type", "application/json")
	switch {
	case call.Type == "form":
		_, _ = w.Write(formData)
		return

	case call.Type == "submit":
		message := "Hello, world!"
		v, ok := call.Values["message"]
		if ok && v != nil {
			message += fmt.Sprintf(" ...and %s!", v)
		}
		mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, message)
	}
	_ = json.NewEncoder(w).Encode(apps.CallResponse{})
}

func sendModal(w http.ResponseWriter, req *http.Request) {
	_, _ = w.Write(formData)
}

func manifest(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(manifestData)
}

func icon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(iconData)
}
