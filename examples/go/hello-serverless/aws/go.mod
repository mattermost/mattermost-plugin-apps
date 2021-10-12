module github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/aws

go 1.16

require (
	github.com/aws/aws-lambda-go v1.19.1
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/mattermost/mattermost-plugin-apps v0.7.1-0.20211011164942-ce8ff2bc1616
	github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function v0.0.0-00010101000000-000000000000
)

replace github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function => ../function
