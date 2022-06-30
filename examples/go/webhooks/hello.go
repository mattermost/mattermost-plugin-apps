package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

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
	Version:     "1.1.0",
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
}

var bindings = []apps.Binding{
	{
		Location: apps.LocationCommand,
		Bindings: []apps.Binding{
			{
				Icon:        "icon.png",
				Label:       "example-webhooks",
				Description: "Example Webhooks App",
				Hint:        "[ subscribe | trigger ]",
				Bindings: []apps.Binding{
					{
						Label:       "subscribe",
						Description: "Subscribes the current channel to the demo webhooks.",
						Submit: apps.NewCall("/subscribe").WithExpand(apps.Expand{
							// Need App to get the app URL and the webhook secret.
							App: apps.ExpandAll.Required(),

							// Need to check the user roles, and act as the user.
							ActingUser:            apps.ExpandAll.Required(),
							ActingUserAccessToken: apps.ExpandAll.Required(),

							// What channel to post webhook to.
							Channel: apps.ExpandAll.Required(),
						}),
					},
					{
						Label:       "trigger",
						Description: "Triggers a demo webhook to the subscribed channel. Make sure you subscribe first.",
						Submit: apps.NewCall("/trigger").WithExpand(apps.Expand{
							// Need App to get the app URL and the webhook secret.
							App: apps.ExpandAll.Required(),
							// Need to check the user roles.
							ActingUser: apps.ExpandAll.Required(),
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

	// `subscribe` command.
	http.HandleFunc("/subscribe", subscribe)

	// `trigger` command.
	http.HandleFunc("/trigger", trigger)

	// Webhook handler
	http.HandleFunc("/webhook/", webhookReceived)

	fmt.Printf("%s app listening on %q \n", manifest.AppID, listenAddr)
	fmt.Printf("Install via /apps install http %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(listenAddr, nil))
}

func webhookURL(cc apps.Context) string {
	u := cc.MattermostSiteURL + cc.AppPath + path.Webhook + "/hello?"
	q := url.Values{}
	q.Set("secret", cc.App.WebhookSecret)
	return u + q.Encode()
}

func subscribe(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	if !creq.Context.ActingUser.IsSystemAdmin() {
		httputils.WriteJSON(w, apps.NewErrorResponse(errors.New("you need to be a system administrator to access the app's webhook secret")))
		return
	}

	// Add the Bot user to the team (if applicable) and the channel.
	asAdmin := appclient.AsActingUser(creq.Context)
	if creq.Context.Channel.TeamId != "" {
		if _, _, err := asAdmin.AddTeamMember(creq.Context.Channel.TeamId, creq.Context.BotUserID); err != nil {
			httputils.WriteJSON(w, apps.NewErrorResponse(err))
			return
		}
	}
	if _, _, err := asAdmin.AddChannelMember(creq.Context.Channel.Id, creq.Context.BotUserID); err != nil {
		httputils.WriteJSON(w, apps.NewErrorResponse(err))
		return
	}

	// Store the channel ID for future use: do it as the Bot user, since it's
	// the Bot user that receives webhooks.
	asBot := appclient.AsBot(creq.Context)
	asBot.KVSet("", "channel_id", creq.Context.Channel.Id)

	asBot.CreatePost(&model.Post{
		ChannelId: creq.Context.Channel.Id,
		Message: fmt.Sprintf(
			"@%s installed me into this channel, I will handle webhooks. "+
				"Try `/hello-webhooks trigger` from anywhere in Mattermost, or POST some data to `%s` to trigger",
			creq.Context.ActingUser.Username, webhookURL(creq.Context)),
	})

	httputils.WriteJSON(w, apps.NewTextResponse("OK"))
}

func trigger(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)
	if !creq.Context.ActingUser.IsSystemAdmin() {
		httputils.WriteJSON(w, apps.NewErrorResponse(errors.New("you need to be a system administrator to access the app's webhook secret")))
		return
	}

	http.Post(webhookURL(creq.Context),
		"application/json",
		bytes.NewReader([]byte(`"Hello from the trigger command!"`)))

	httputils.WriteJSON(w,
		apps.NewTextResponse("posted a Hello webhook message in the subscribed channel"))
}

func webhookReceived(w http.ResponseWriter, req *http.Request) {
	creq := apps.CallRequest{}
	json.NewDecoder(req.Body).Decode(&creq)

	asBot := appclient.AsBot(creq.Context)
	channelID := ""
	asBot.KVGet("", "channel_id", &channelID)

	asBot.CreatePost(&model.Post{
		ChannelId: channelID,
		Message:   fmt.Sprintf("received webhook, path `%s`, data: `%v`", creq.Path, creq.Values["data"]),
	})

	httputils.WriteJSON(w, apps.NewTextResponse("OK"))
}
