package main

import (
	mattermost "github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/mattermost/mattermost-plugin-cloudapps/server/configurator"
	"github.com/mattermost/mattermost-plugin-cloudapps/server/plugin"
)

var BuildHash string
var BuildHashShort string
var BuildDate string

func main() {
	mattermost.ClientMain(
		plugin.NewPlugin(
			&configurator.BuildConfig{
				Manifest:       manifest,
				BuildHash:      BuildHash,
				BuildHashShort: BuildHashShort,
				BuildDate:      BuildDate,
			}))
}
