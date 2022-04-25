module github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/aws

go 1.16

require (
	github.com/aws/aws-lambda-go v1.19.1
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/mattermost/mattermost-plugin-apps v1.0.0
	github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/function v0.0.0-00010101000000-000000000000
)

replace github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/function => ../function
