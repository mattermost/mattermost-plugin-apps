# Hello World!

Here is an example of an HTTP App ([source](/server/examples/go/helloworld)),
written in Go and runnable on http://localhost:8080. 

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

Then you can try clicking the "Hello World" channel header button, which brings up a modal:
![image](https://user-images.githubusercontent.com/1187448/110829345-da81d800-824c-11eb-96e7-c62637242897.png)
type `testing` and click Submit, you should see:
![image](https://user-images.githubusercontent.com/1187448/110829449-fb4a2d80-824c-11eb-8ade-d20e0fbd1b94.png)

You can also use `/helloworld send` command.

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
