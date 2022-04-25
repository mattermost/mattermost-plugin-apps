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
	rootURL    = "http://localhost:8081"
	listenAddr = ":8081"
)

//go:embed icon.png
var iconData []byte

var manifest = apps.Manifest{
	AppID:       "example-webhooks",
	Version:     "1.0.0",
	DisplayName: "Example of an app receiving webhooks.",
	Icon:        "icon.png",
	HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/webhooks",
	RequestedPermissions: apps.Permissions{
		apps.PermissionActAsUser,
		apps.PermissionActAsBot,
		apps.PermissionRemoteWebhooks,
	},
	RequestedLocations: apps.Locations{
		apps.LocationCommand,
	},
	Deploy: apps.Deploy{
		HTTP: &apps.HTTP{
			RootURL: rootURL,
		},
	},
	RemoteWebhookAuthType: apps.SecretAuth,
	OnInstall: apps.NewCall("/install").WithExpand(apps.Expand{
		ActingUserAccessToken: apps.ExpandAll,
	}),
}

var bindings = []apps.Binding{
	{
		Location: apps.LocationCommand,
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "example-webhooks",
				Description: "Example Webhooks App",
				Hint:        "[ send ]",
				Bindings: []apps.Binding{
					{
						Label: "send",
						Form: &apps.Form{
							Title:  "Send a test webhook message",
							Icon:   "icon.png",
							Submit: apps.NewCall("/send"),
							Fields: []apps.Field{
								{
									Name:                 "url",
									Type:                 "text",
									IsRequired:           true,
									AutocompletePosition: 1,
								},
							},
						},
					},
					{
						Label: "info",
						Submit: apps.NewCall("/info").WithExpand(apps.Expand{
							App: apps.ExpandAll,
						}),
					},
				},
			},
		},
	},
}

func main() {
	http.HandleFunc("/manifest.json", httputils.DoHandleJSON(manifest))
	http.HandleFunc("/bindings", httputils.DoHandleJSON(apps.NewDataResponse(bindings)))
	http.HandleFunc("/static/icon.png", httputils.DoHandleData("image/png", iconData))

	// install handler - uses the admin token to allow the bot to post to
	// current channel.
	http.HandleFunc("/install", install)

	// Webhook handler
	http.HandleFunc("/webhook/", webhookReceived)

	// `info` command - displays the webhook URL.
	http.HandleFunc("/info", info)

	// `send` command - send a test webhook message.
	http.HandleFunc("/send", send)

	fmt.Printf("%s app listening on %q \n", manifest.AppID, listenAddr)
	fmt.Printf("Install via /apps install http %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(listenAddr, nil))
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
		Message:   fmt.Sprintf("@%s is installed into this channel, try /%s-webhooks send", manifest.AppID, manifest.AppID),
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
		apps.NewTextResponse("Try `/%s send %s/anything?secret=%s`",
			manifest.AppID,
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
		bytes.NewReader([]byte(`"A text webhook message"`)))

	httputils.WriteJSON(w,
		apps.NewTextResponse("posted a test webhook message"))
}
