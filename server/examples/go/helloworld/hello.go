package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
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
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Returns the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// The form for sending a Hello message.
	http.HandleFunc("/send/form", sendForm)

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/submit", sendSubmit)

	// The main handler for sending a Hello message.
	http.HandleFunc("/send/lookup", sendLookup)

	// Forces the send form to be displayed as a modal.
	// TODO: ticket: this should be unnecessary.
	http.HandleFunc("/send-modal/submit", writeJSON(formData))

	// Serves the icon for the App.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	fmt.Println("listening")
	http.ListenAndServe(":8080", nil)
}

func sendSubmit(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	message := "Hello, world!"
	v, _ := c.Values["message"].(string)
	if v == "cause an error" {
		data := map[string]interface{}{
			"errors": map[string]string{
				"message": "This field seems to have an invalid value.",
			},
		}
		resp := apps.CallResponse{
			Type: apps.CallResponseTypeError,
			Data: data,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if v != "" {
		message += fmt.Sprintf(" ...and %s!", v)
	}

	recipient := c.Context.ActingUserID
	recipientLabel := "you"
	user, _ := (c.Values["user"]).(map[string]interface{})
	if user != nil {
		recipient, _ = user["value"].(string)
		recipientLabel, _ = user["label"].(string)
	}

	mmclient.AsBot(c.Context).DM(recipient, message)

	resp := apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: md.Markdownf("Sent a survey to %s.", recipientLabel),
	}
	json.NewEncoder(w).Encode(resp)
}

func sendLookup(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	items := []apps.SelectOption{
		{
			Label: "Option 1",
			Value: "option1",
		},
		{
			Label: "Option 2",
			Value: "option2",
		},
	}
	data := map[string]interface{}{
		"items": items,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps.CallResponse{Type: apps.CallResponseTypeOK, Data: data})
}

func sendForm(w http.ResponseWriter, req *http.Request) {
	c := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&c)

	resp := &apps.CallResponse{}
	json.Unmarshal(formData, &resp)

	resp.Form = populateForm(resp.Form, c.Values)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func populateForm(form *apps.Form, values map[string]interface{}) *apps.Form {
	if form == nil || values == nil {
		return form
	}

	for name, value := range values {
		for _, field := range form.Fields {
			if name == field.Name {
				field.Value = value
			}
		}
	}

	return form
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
