package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

//go:embed manifest.json
var manifestData []byte

const (
	host = "localhost"
	port = 8083
)

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", httputils.HandleStaticJSONData(manifestData))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings", httputils.HandleStaticJSONData([]byte("{}")))

	http.HandleFunc("/install", respondWithMessage("Thanks for installing me!"))

	http.HandleFunc("/uninstall", respondWithMessage("No, don't uninstall me!"))

	http.HandleFunc("/enable", respondWithMessage("I'm back up again"))

	http.HandleFunc("/disable", respondWithMessage("Takeing a little nap"))

	addr := fmt.Sprintf(":%v", port)
	rootURL := fmt.Sprintf("http://%v:%v", host, port)
	fmt.Printf("hello-lifecycle app listening on %q \n", addr)
	fmt.Printf("Install via /apps install url %s/manifest.json \n", rootURL)
	panic(http.ListenAndServe(addr, nil))
}

func respondWithMessage(message string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		c := apps.CallRequest{}
		json.NewDecoder(req.Body).Decode(&c)

		_, err := appclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)
		if err != nil {
			json.NewEncoder(w).Encode(apps.NewErrorResponse(err))
			return
		}

		httputils.WriteJSON(w,
			apps.NewTextResponse("Created a post in your DM channel."))
	}
}
