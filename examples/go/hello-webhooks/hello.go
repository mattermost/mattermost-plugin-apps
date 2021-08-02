package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/server"
	"github.com/mattermost/mattermost-plugin-apps/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

//go:embed icon.png
var iconData []byte

//go:embed manifest.json
var manifestData []byte

//go:embed bindings.json
var bindingsData []byte

//go:embed info_form.json
var infoFormData []byte

//go:embed send_form.json
var sendFormData []byte

func main() {
	// Static handlers

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", writeJSON(bindingsData))

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png", writeData("image/png", iconData))

	// install handler
	http.HandleFunc("/install", install)

	// Webhook handler
	http.HandleFunc("/webhook/", webhookReceived)

	// `info` command - displays the webhook URL.
	http.HandleFunc("/info/form", writeJSON(infoFormData))
	http.HandleFunc("/info/submit", info)

	// `send` command - send a Hello webhook message.
	http.HandleFunc("/send/form", writeJSON(sendFormData))
	http.HandleFunc("/send/submit", send)

	server.Run(manifestData)
}

func install(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	teamID := creq.Context.TeamID
	channelID := creq.Context.ChannelID

	// Add the Bot user to the team and the channel.
	asAdmin := mmclient.AsAdmin(creq.Context)
	asAdmin.AddTeamMember(teamID, creq.Context.BotUserID)
	asAdmin.AddChannelMember(channelID, creq.Context.BotUserID)

	asBot := mmclient.AsBot(creq.Context)
	// store the channel ID for future use
	asBot.KVSet("channel_id", "", channelID)

	asBot.CreatePost(&model.Post{
		ChannelId: channelID,
		Message:   "@hello-webhooks is installed into this channel, try /hello-webhooks send",
	})

	json.NewEncoder(w).Encode(apps.CallResponse{Type: apps.CallResponseTypeOK})
}

func webhookReceived(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	asBot := mmclient.AsBot(creq.Context)
	channelID := ""
	asBot.KVGet("channel_id", "", &channelID)

	asBot.CreatePost(&model.Post{
		ChannelId: channelID,
		Message:   fmt.Sprintf("received webhook, path `%s`, data: `%v`", creq.Path, creq.Values["data"]),
	})

	json.NewEncoder(w).Encode(apps.CallResponse{Type: apps.CallResponseTypeOK})
}

func info(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: md.Markdownf("Try `/hello-webhooks send %s`",
			creq.Context.MattermostSiteURL+creq.Context.AppPath+apps.PathWebhook+
				"/hello"+
				"?secret="+creq.Context.App.WebhookSecret),
	})
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	url, _ := creq.Values["url"].(string)

	http.Post(
		url,
		"application/json",
		bytes.NewReader([]byte(`"Hello from a webhook!"`)))

	json.NewEncoder(w).Encode(apps.CallResponse{
		Markdown: "posted a Hello webhook message",
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
