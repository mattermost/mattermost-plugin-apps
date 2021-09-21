package hello_serverless

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

//go:embed manifest.json
var manifestData []byte
var m apps.Manifest

func init() {
	err := json.Unmarshal(manifestData, &m)
	if err != nil {
		panic(err)
	}

	// Serve the app's manifest.
	http.HandleFunc("/manifest.json",
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(manifestData)
		})
}

func Manifest() apps.Manifest {
	return m
}
