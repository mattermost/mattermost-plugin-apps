package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

//go:embed manifest.json
var manifestData []byte

const (
	host = "localhost"
	port = 8080
)

func main() {
	// Serve its own manifest as HTTP for convenience in dev. mode.
	http.HandleFunc("/manifest.json", writeJSON(manifestData))

	// Returns the Channel Header and Command bindings for the app.
	http.HandleFunc("/bindings", writeJSON([]byte("{}")))

	http.HandleFunc("/install", respondWithMessage("Thanks for installing me!"))

	http.HandleFunc("/uninstall", respondWithMessage("No, don't uninstall me!"))

	http.HandleFunc("/enable", respondWithMessage("I'm back up again"))

	http.HandleFunc("/disable", respondWithMessage("Takeing a little nap"))

	addr := fmt.Sprintf("%v:%v", host, port)
	rootURL := "http://" + addr
	fmt.Printf("hello-lifecycle app listening at %s \n", rootURL)
	fmt.Printf("Install via /apps install url %s/manifest.json \n", rootURL)
	http.ListenAndServe(addr, nil)
}

func respondWithMessage(message string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		c := apps.CallRequest{}
		json.NewDecoder(req.Body).Decode(&c)

		_, err := mmclient.AsBot(c.Context).DM(c.Context.ActingUserID, message)
		if err != nil {
			json.NewEncoder(w).Encode(apps.CallResponse{
				Type:      apps.CallResponseTypeError,
				ErrorText: err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(apps.CallResponse{
			Type:     apps.CallResponseTypeOK,
			Markdown: "Created a post in your DM channel.",
		})
	}
}

func writeData(ct string, data []byte) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", ct)
		w.Write(data)
	}
}

func writeJSON(data []byte) func(w http.ResponseWriter, r *http.Request) {
	return writeData("application/json", data)
}
