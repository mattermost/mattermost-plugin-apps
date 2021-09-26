package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/hello-serverless/function"
)

// OpenFaaS provides its own main via the golang-middleware template, so it's
// not explicitly supported here.
func main() {
	function.InitApp(apps.DeployAWSLambda)
	lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
}
