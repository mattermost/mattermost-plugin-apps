package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed icon.png
var iconData []byte

func main() {
	http.HandleFunc("/manifest.json", manifest)
	http.HandleFunc("/bindings", bindings)

	http.HandleFunc("/hello", hello)
	http.HandleFunc("/static/icon.png", icon)

	http.ListenAndServe(":8080", nil)
}

func manifest(w http.ResponseWriter, req *http.Request) {
	m := apps.Manifest{
		AppID:       "helloworld",
		DisplayName: "Hello, world!",
		Type:        apps.AppTypeHTTP,
		HTTPRootURL: "http://localhost:8080",
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsBot,
		},
		RequestedLocations: apps.Locations{
			apps.LocationChannelHeader,
			apps.LocationCommand,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(m)
}

func bindings(w http.ResponseWriter, req *http.Request) {
	hello := &apps.Call{
		Path: "/hello",
	}
	bindings := []*apps.Binding{
		{
			Location: apps.LocationChannelHeader,
			Bindings: []*apps.Binding{
				{
					Location: "message",
					//TODO: ticket: relative URL doesn't work
					Icon: "http://localhost:8080/static/icon.png",
					Call: hello,
				},
			},
		},
		{
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					Location: "message",
					//TODO: ticket: Location alone does not work, requires Label
					Label: "message",
					Call:  hello,
				},
			},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: bindings,
	})
}

func hello(w http.ResponseWriter, req *http.Request) {
	call := apps.Call{}
	_ = json.NewDecoder(req.Body).Decode(&call)

	w.Header().Set("Content-Type", "application/json")
	switch call.Type {
	case apps.CallTypeForm:
		_ = json.NewEncoder(w).Encode(apps.CallResponse{
			//TODO: ticket: client is erroring with `App response type was not expected. Response type: ok.` if {} is returned.
			Type: apps.CallResponseTypeForm,
			//TODO: ticket: client is erroring with `Response type is form, but no form was included in response.` if not initialized.
			Form: &apps.Form{},
		})
		return

	case apps.CallTypeSubmit:
		mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, "Hello, world!")

	}

	_ = json.NewEncoder(w).Encode(apps.CallResponse{})
}

func icon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, bytes.NewReader(iconData))
}
