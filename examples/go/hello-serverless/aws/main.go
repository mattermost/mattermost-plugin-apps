package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function"
)

func main() {
	function.DeployType = apps.DeployAWSLambda
	lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
}
