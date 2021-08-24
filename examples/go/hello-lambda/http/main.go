package main

import (
	app "github.com/mattermost/mattermost-plugin-apps/examples/go/hello-lambda"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/server"
)

func main() {
	server.Run(app.ManifestData)
}
