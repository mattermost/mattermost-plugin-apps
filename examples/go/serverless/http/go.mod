module github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/http

go 1.16

require (
	github.com/mattermost/mattermost-plugin-apps v1.0.0
	github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/function v0.0.0-00010101000000-000000000000
)

replace github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/function => ../function
