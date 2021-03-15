package main

import (
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

var BuildHash string
var BuildHashShort string
var BuildDate string

func main() {
	plugin.ClientMain(
		NewPlugin(
			&config.BuildConfig{
				Manifest:       manifest,
				BuildHash:      BuildHash,
				BuildHashShort: BuildHashShort,
				BuildDate:      BuildDate,
			}))
}
