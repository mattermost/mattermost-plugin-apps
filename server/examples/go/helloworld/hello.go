package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed icon.png
var iconData []byte

func main() {
	http.HandleFunc("/", catchAll)
	http.HandleFunc("/manifest.json", manifest)
	http.HandleFunc("/bindings", bindings)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/static/icon.png", icon)
	http.ListenAndServe(":8080", nil)
}

func catchAll(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("<><> catchAll: %s\n", r.URL.String())
}

func icon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, bytes.NewReader(iconData))
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
	fmt.Printf("<><> 1: %s\n", call.Type)

	w.Header().Set("Content-Type", "application/json")
	switch call.Type {
	case apps.CallTypeForm:
		_ = json.NewEncoder(w).Encode(apps.CallResponse{
			//TODO: ticket: client is erroring with `App response type not supported. Response type: {type}.` if {} is returned.
			Type: apps.CallResponseTypeForm,
			//TODO: ticket: client is erroring with `Response type is form, but no form was included in response.` if not initialized.
			Form: &apps.Form{},
		})
		return

	case apps.CallTypeSubmit:
		fmt.Printf("<><> 2: DM to %s\n", call.Context.ActingUserID)
		mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, "Hello, world!")

	}

	_ = json.NewEncoder(w).Encode(apps.CallResponse{
		//TODO: ticket: OK should be defaulted on by the proxy, {} should be enough
		Type: apps.CallResponseTypeOK,
	})
}
