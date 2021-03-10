# Overview

Apps are lighweight interactive add-ons to mattermost. Apps can:
- display interactive, dynamic Modal forms and Message Actions.
- attach themselves to locations in the Mattermost UI (e.g. channel bar buttons,
  post menu, channel menu, commands), and can add their custom /commands with
  full Autocomplete.
- receive webhooks from Mattermost, and from 3rd parties, and use the Mattermost
  REST APIs to post messages, etc. 
- be hosted externally (HTTP), on Mattermost Cloud (AWS Lambda), and soon
  on-prem and in customers' own AWS environments.
- be developed in any language*

# Hello World!
Here is an example of an HTTP App, written in Go. It adds a channel header
button, and a command to send "Hello" messages. See
/server/examples/go/helloworld. 

To install Hello, World follow these steps,
- cd .../mattermost-plugin-apps/server/examples/go/helloworld
- `go run .` - note go 1.16 is required
- In Mattermost, `/apps debug-add-manifest --url http://localhost:8080/manifest.json`
  and `/apps install --app-id helloworld`

Then you can try clicking the "Hello World" channel header button, or using
`/helloworld send` command.


There are 4 principal pieces to the App: `manifest`, `bindings` handler,
functions (`send`, `send-modal`), and the icon.

```go
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
```

### Manifest
The manifest declares App metadata, and for AWS Lambda apps declares the Call
Path to Lambda Function mappings. For HTTP apps, paths are prefixed with
HTTPRootURL before invoking, so no mappings are needed.

The Hello World App is an HTTP app. It requests the permission to act as a Bot,
and to add UI to the channel header, and to /commands.

```go
func manifest(w http.ResponseWriter, req *http.Request) {
	m := apps.Manifest{
		AppID:                "helloworld",
		DisplayName:          "Hello, world!",
		Type:                 "http",
		HTTPRootURL:          "http://localhost:8080",
		RequestedPermissions: apps.Permissions{"act_as_bot"},
		RequestedLocations:   apps.Locations{"/channel_header", "/command"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(m)
}
```

### Bindings and Locations
Locations are named elements in Mattermost UI. Bindings specify how App's calls
should be displayed at, and invoked from these locations. 

The Hello App creates a Channel Header button, and adds a `/helloworld send` command.

```go
func bindings(w http.ResponseWriter, req *http.Request) {
	bindings := []*apps.Binding{
		{
			Location: "/channel_header",
			Bindings: []*apps.Binding{
				{
					Location: "send-button",
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
					Location: "send",
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
```

### Functions
Functions handle user events and webhooks. The Hello World App exposes 2 functions:
- `/send` that services the command and modal
- `/send-modal` that forces the modal to be displayed

### Icons 
Apps may include static assets. At the moment, only icons are used.

### OAuth2 support
Apps rely on user-level OAuth2 authentication to impersonate Mattermost users,
or to execute administrative tasks. Apps are expected to rely on user-level
OAuth2 to 3rd party systems.

## Development environment

## Forms and Functions
### Authentication
### Context Expansion
### Call vs Notification

## Interactive Features
### Locations and Bindings
### /commands and autocomplete
### Modals
### In-Post interactivity

## Using Mattermost APIs
### Authentication and Token Store
### Scopes and Permissions
### Apps Subscriptions API
### Apps KV API
### Mattermost REST API

## Using 3rd party APIs
### Authentication and Token Store
### 3rd party webhooks

## Lifecycle
### Development
### Submit to Marketplace
### Provision
### Publish
### Install
### Uninstall
### Upgrade/downgrade consideration