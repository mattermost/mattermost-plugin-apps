module github.com/mattermost/mattermost-example-apps/hello-serverless

go 1.16

require (
	github.com/aws/aws-lambda-go v1.19.1
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/mattermost/mattermost-plugin-apps v0.7.1-0.20210914085138-b6c7743e6a75
)

replace github.com/mattermost/mattermost-example-apps/hello-serverless/app => ./app