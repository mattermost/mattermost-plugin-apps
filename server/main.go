package main

import (
	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

var BuildDate string
var BuildHash string
var BuildHashShort string

func main() {
	plugin.ClientMain(
		NewPlugin(
			config.BuildConfig{
				Manifest:       manifest,
				BuildHash:      BuildHash,
				BuildHashShort: BuildHashShort,
				BuildDate:      BuildDate,
			}))
}
