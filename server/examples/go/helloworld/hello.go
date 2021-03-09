package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
)

func main() {
	http.HandleFunc("/manifest.json", manifest)
	http.HandleFunc("/bindings", bindings)
	http.HandleFunc("/hello", hello)
	http.ListenAndServe(":8080", nil)
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func manifest(w http.ResponseWriter, req *http.Request) {
	httputils.WriteJSON(w,
		apps.Manifest{
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
		})
}

func bindings(w http.ResponseWriter, req *http.Request) {
	hello := &apps.Call{
		Path: "/hello",
	}

	httputils.WriteJSON(w, []*apps.Binding{
		{
			Location: apps.LocationChannelHeader,
			Bindings: []*apps.Binding{
				{
					Location: "hello",
					Call:     hello,
				},
			},
		}, {
			Location: apps.LocationCommand,
			Bindings: []*apps.Binding{
				{
					Location: "hello",
					Call:     hello,
				},
			},
		},
	})
}

func hello(w http.ResponseWriter, req *http.Request) {
	call := apps.Call{}
	_ = json.NewDecoder(req.Body).Decode(&call)

	mmclient.AsBot(call.Context).DM(call.Context.ActingUserID, "Hello, world!")
}
