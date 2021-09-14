package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"

	function "github.com/mattermost/mattermost-example-apps/hello-serverless/app"
)

// OpenFaaS provides its own main via the golang-middleware template, so it's
// not explicitly supported here.
func main() {
	deployType := apps.DeployType(os.Getenv("MODE"))
	if deployType == "" {
		deployType = apps.DeployAWSLambda
	}
	function.InitApp(deployType)
	switch deployType {
	// AWS Lambda
	case apps.DeployAWSLambda:
		lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
	case apps.DeployHTTP:
		fmt.Println("Listening on :8080")
		panic(http.ListenAndServe(":8080", nil))
	default:
		panic(deployType.String() + " is not supported")
	}
}
