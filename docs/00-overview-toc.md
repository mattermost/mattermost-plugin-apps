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
Here is an example of an HTTP App, written in Go and runnable on
http://localhost:8080. [Source](/server/examples/go/helloworld)

- In its `manifest.json` it declares itself an HTTP application.
- It contains a `send` function that sends a parameterized message back to the
  user. 
- It contains a `send-modal` function that forces displaying the `send` form as
  a Modal.
- In its `bindings` function it attaches `send-modal` to a button in the channel
  header, and `send` to a /helloworld command

To install "Hello, World" on a locally-running instance of Mattermost follow
these steps (go 1.16 is required):
```sh
cd .../mattermost-plugin-apps/server/examples/go/helloworld
go run . 
```

In Mattermost desktop app run:
```
/apps debug-add-manifest --url http://localhost:8080/manifest.json
/apps install --app-id helloworld
```

Then you can try clicking the "Hello World" channel header button, or using
`/helloworld send` command.

## Manifest
The manifest declares App metadata, and for AWS Lambda apps declares the Call
Path to Lambda Function mappings. For HTTP apps, paths are prefixed with
HTTPRootURL before invoking, so no mappings are needed.

The Hello World App is an HTTP app. It requests the permission to act as a Bot,
and to add UI to the channel header, and to /commands.

```json
{
	"app_id": "helloworld",
	"display_name": "Hello, world!",
	"app_type": "http",
	"root_url": "http://localhost:8080",
	"requested_permissions": [
		"act_as_bot"
	],
	"requested_locations": [
		"/channel_header",
		"/command"
	]
}
```

## Bindings and Locations
Locations are named elements in Mattermost UI. Bindings specify how App's calls
should be displayed at, and invoked from these locations. 

The Hello App creates a Channel Header button, and adds a `/helloworld send` command.

```json
{
	"type": "ok",
	"data": [
		{
			"location": "/channel_header",
			"bindings": [
				{
					"location": "send-button",
					"icon": "http://localhost:8080/static/icon.png",
					"call": {
						"path": "/send-modal"
					}
				}
			]
		},
		{
			"location": "/command",
			"bindings": [
				{
					"location": "send",
					"label": "send",
					"call": {
						"path": "/send"
					}
				}
			]
		}
	]
}
```

## Functions and Form
Functions handle user events and webhooks. The Hello World App exposes 2 functions:
- `/send` that services the command and modal.
- `/send-modal` that forces the modal to be displayed.

```go
func send(w http.ResponseWriter, req *http.Request) {
	call := apps.Call{}
	out := apps.CallResponse{}

	_ = json.NewDecoder(req.Body).Decode(&call)
	w.Header().Set("Content-Type", "application/json")
	switch {
	case call.Type == "form":
		out = helloForm()

	case call.Type == "submit":
		message := "Hello, world!"
		v, ok := call.Values["message"]
		if ok && v != nil {
			message += fmt.Sprintf(" ...and %s!", v)
		}
		mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, message)
	}
	_ = json.NewEncoder(w).Encode(out)
}

func sendModal(w http.ResponseWriter, req *http.Request) {
	_ = json.NewEncoder(w).Encode(helloForm())
}
```

The functions use a simple form with 1 text field named `"message"`, the form
submits to `/send`.

```json
{
	"type": "form",
	"form": {
		"title": "Hello, world!",
		"icon": "http://localhost:8080/static/icon.png",
		"fields": [
			{
				"type": "text",
				"name": "message",
				"label": "message"
			}
		],
		"call": {
			"path": "/send"
		}
	}
}
```

## Icons 
Apps may include static assets. At the moment, only icons are used.

## OAuth2 support
Apps rely on user-level OAuth2 authentication to impersonate Mattermost users,
or to execute administrative tasks. Apps are expected to rely on user-level
OAuth2 to 3rd party systems.

# Development environment
See https://docs.google.com/document/d/1-o9A8l65__rYbx6O-ZdIgJ7LJgZ1f3XRXphAyD7YfF4/edit#

# Functions
## Call
## Authentication
## Context Expansion
## Special Notes
### Use of router packages in Apps
- Go (gorilla mux)
- JavaScript
### Call vs Notification
### AWS Lambda packaging

# Forms
## Binding Forms to Locations
### Channel Header
### Post Menu
### /Command Autocomplete
## Autocomplete
## Modals


# In-Post Interactivity

# Using Mattermost APIs
## Authentication and Token Store
## Scopes and Permissions
## Apps Subscriptions API
## Apps KV API
## Mattermost REST API

# Using 3rd party APIs
## Authentication and Token Store
## 3rd party webhooks

# App Lifecycle
## Development
## Submit to Marketplace
## Provision
## Publish
## Install
## Uninstall
## Upgrade/downgrade consideration