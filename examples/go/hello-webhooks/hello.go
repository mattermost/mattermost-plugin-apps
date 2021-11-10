package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/mattermost/mattermost-server/v6/model"
)

const (
	host = "localhost"
	port = 8081
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

	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.HandleStaticJSONData(manifestData))

	// Serve the Channel Header and Command bindings for the App.
	http.HandleFunc("/bindings", httputils.HandleStaticJSONData(bindingsData))

	// Serve the icon for the App.
	http.HandleFunc("/static/icon.png",
		httputils.HandleStaticData("image/png", iconData))

	// install handler
	http.HandleFunc("/install", install)

	// Webhook handler
	http.HandleFunc("/webhook/", webhookReceived)

	// `info` command - displays the webhook URL.
	http.HandleFunc("/info/form", httputils.HandleStaticJSONData(infoFormData))
	http.HandleFunc("/info/submit", info)

	// `send` command - send a Hello webhook message.
	http.HandleFunc("/send/form", httputils.HandleStaticJSONData(sendFormData))
	http.HandleFunc("/send/submit", send)

	addr := fmt.Sprintf(":%v", port)
	rootURL := fmt.Sprintf("http://%v:%v", host, port)
	fmt.Printf("hello-webhooks app listening on %q \n", addr)
	fmt.Printf("Install via /apps install http %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(addr, nil))
}

func install(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	teamID := creq.Context.TeamID
	channelID := creq.Context.ChannelID

	// Add the Bot user to the team and the channel.
	asAdmin := appclient.AsActingUser(creq.Context)
	asAdmin.AddTeamMember(teamID, creq.Context.BotUserID)
	asAdmin.AddChannelMember(channelID, creq.Context.BotUserID)

	asBot := appclient.AsBot(creq.Context)
	// store the channel ID for future use
	asBot.KVSet("channel_id", "", channelID)

	asBot.CreatePost(&model.Post{
		ChannelId: channelID,
		Message:   "@hello-webhooks is installed into this channel, try /hello-webhooks send",
	})

	httputils.WriteJSON(w, apps.NewTextResponse("OK"))
}

func webhookReceived(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	asBot := appclient.AsBot(creq.Context)
	channelID := ""
	asBot.KVGet("channel_id", "", &channelID)

	asBot.CreatePost(&model.Post{
		ChannelId: channelID,
		Message:   fmt.Sprintf("received webhook, path `%s`, data: `%v`", creq.Path, creq.Values["data"]),
	})

	httputils.WriteJSON(w, apps.NewTextResponse("OK"))
}

func info(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	httputils.WriteJSON(w,
		apps.NewTextResponse("Try `/hello-webhooks send %s/hello?secret=%s`",
			creq.Context.MattermostSiteURL+creq.Context.AppPath+path.Webhook,
			creq.Context.App.WebhookSecret))
}

func send(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	url, _ := creq.Values["url"].(string)

	http.Post(
		url,
		"application/json",
		bytes.NewReader([]byte(`"Hello from a webhook!"`)))

	httputils.WriteJSON(w,
		apps.NewTextResponse("posted a Hello webhook message"))
}
