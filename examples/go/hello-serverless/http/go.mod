module github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/http

go 1.16

require (
	github.com/mattermost/mattermost-plugin-apps v0.7.1-0.20211207113443-6e904255aa08
	github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function v0.0.0-00010101000000-000000000000
)

replace github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function => ../function
