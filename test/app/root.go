package main

import (
	"embed" // Need to embed manifest file
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// appManifestData is preloaded with the Mattermost App manifest.
//go:embed manifest.json
var appManifestData []byte

// StaticFS is preloaded with the contents of the ./static directory.
//go:embed static
var StaticFS embed.FS

var AppManifest apps.Manifest

func init() {
	err := json.Unmarshal(appManifestData, &AppManifest)
	if err != nil {
		panic(err)
	}
}
