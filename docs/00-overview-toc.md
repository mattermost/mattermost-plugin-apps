## Overview
### What are Apps?
- Apps are lighweight interactive add-ons to mattermost. 
- Apps can display interactive, dynamic Modal forms.
- Apps can attach themselves to locations in the Mattermost UI (e.g. channel bar buttons, post menu, channel menu, commands), and can add their custom /commands with full Autocomplete.
- Apps can receive webhooks from Mattermost, and from 3rd parties, and use the Mattermost REST APIs to post messages, etc. 
- Apps can be hosted externally (HTTP), on Mattermost Cloud (AWS Lambda), and soon on-prem and in customers' own AWS environments.
- Apps can be developed in any language*

## Anatomy of an App: Hello World!
Adds a channel header button, and a command to send "Hello" messages. See /server/examples/go/helloworld. A standalone HTTP app, running on http://localhost:8080.


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

The Hello World App requests the permission to act as a Bot, and to add UI to
the channel header, and to /commands.

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
When the App's channel header button is clicked, or `/helloworld` command is executed, a **Call** is made to the App's **function** matching the call path. The App can then perform its task, or respond with a **Form** to gather more data from the user.

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