package main

import (
	root "github.com/mattermost/mattermost-plugin-apps"

	"github.com/mattermost/mattermost/server/public/plugin"
)

var manifest = root.Manifest

func main() {
	plugin.ClientMain(NewPlugin(manifest))
}
