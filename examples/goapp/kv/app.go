package main

import (
	"embed"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// static is preloaded with the contents of the ./static directory.
//go:embed static
var static embed.FS

// main starts the app, as a standalone HTTP server. Use $PORT and $ROOT_URL to
// customize.
func main() {
	// Create the app, add `send` to the app's command and the channel header.
	goapp.MakeAppOrPanic(
		apps.Manifest{
			AppID:       "example-kv",
			Version:     "v1.2.0",
			DisplayName: "Example of using the KV store",
			Icon:        "icon.png",
			HomepageURL: "https://github.com/mattermost/mattermost-plugin-apps/examples/go/goapp",
			RequestedPermissions: []apps.Permission{
				apps.PermissionActAsBot,
				apps.PermissionActAsUser,
			},
		},
		goapp.WithStatic(static),
		goapp.WithCommand(get, set),
	).RunHTTP()
}

var set = goapp.MakeBindableFormOrPanic("set",
	apps.Form{
		Title: "Store a value in the KV store",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Name:          "prefix",
				Description:   "The namespace prefix to use, just 2 charachters, don't even ask why...",
				TextMaxLength: 2,
			},
			{
				Name:          "key",
				Description:   "The key (id) to use",
				TextMaxLength: 28,
			},
			{
				Name:        "value",
				Description: "The value to store, as text",
			},
			{
				Name:        "as_bot",
				Description: "Act as the app's bot, as opposed to the acting user",
				Type:        apps.FieldTypeBool,
			},
		},
		Submit: &apps.Call{
			Expand: &apps.Expand{
				ActingUser:            apps.ExpandID.Required(),
				ActingUserAccessToken: apps.ExpandAll.Required(),
			},
		},
	},
	func(creq goapp.CallRequest) apps.CallResponse {
		prefix := creq.GetValue("prefix", "")
		key := creq.GetValue("key", "")
		value := creq.GetValue("value", "")

		client := creq.AsBot()
		asBot, _ := creq.BoolValue("as_bot")
		if !asBot {
			client = creq.AsActingUser()
		}
		changed, err := client.KVSet(prefix, key, value)
		if err != nil {
			return apps.NewTextResponse("Error: %v", err)
		}
		return apps.NewTextResponse("Stored a value in the KV store: prefix: %q, key: %q, value: %q, changed: %v", prefix, key, value, changed)
	},
)

var get = goapp.MakeBindableFormOrPanic("get",
	apps.Form{
		Title: "Get a value from the KV store",
		Icon:  "icon.png",
		Fields: []apps.Field{
			{
				Name:          "prefix",
				Description:   "The namespace prefix to use, just 2 charachters, don't even ask why...",
				TextMaxLength: 2,
			},
			{
				Name:          "key",
				Description:   "The key (id) to use",
				TextMaxLength: 28,
			},
			{
				Name:        "as_bot",
				Description: "Act as the app's bot, as opposed to the acting user",
				Type:        apps.FieldTypeBool,
			},
		},
		Submit: &apps.Call{
			Expand: &apps.Expand{
				ActingUser:            apps.ExpandID.Required(),
				ActingUserAccessToken: apps.ExpandAll.Required(),
			},
		},
	},
	func(creq goapp.CallRequest) apps.CallResponse {
		prefix := creq.GetValue("prefix", "")
		key := creq.GetValue("key", "")

		client := creq.AsBot()
		asBot, _ := creq.BoolValue("as_bot")
		if !asBot {
			client = creq.AsActingUser()
		}

		var value interface{}
		err := client.KVGet(prefix, key, &value)
		if err != nil {
			return apps.NewTextResponse("Error: %v", err)
		}
		return apps.NewTextResponse("Read a value from the KV store: prefix: %q, key: %q, value: %#v", prefix, key, value)
	},
)
