module github.com/mattermost/mattermost-example-apps/hello-serverless

go 1.16

require (
	github.com/aws/aws-lambda-go v1.26.0
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/mattermost/mattermost-example-apps/hello-serverless/app v0.0.0-00010101000000-000000000000
	github.com/mattermost/mattermost-plugin-apps v0.8.0
)

replace (
	github.com/mattermost/mattermost-example-apps/hello-serverless/app => ./app
	github.com/mattermost/mattermost-plugin-apps => github.com/mattermost/mattermost-plugin-apps v0.7.1-0.20210828171049-9c91624214f7
)
