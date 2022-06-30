package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/examples/go/serverless/hello"
)

func main() {
	deployType := apps.DeployType(os.Getenv(hello.DEPLOY_TYPE))
	if deployType == "" {
		deployType = apps.DeployAWSLambda
	}
	hello.DeployType = deployType

	switch deployType {
	case apps.DeployHTTP:
		fmt.Printf("hello-oauth2 app listening on :8080\n")
		fmt.Printf("Install via /apps install http http://localhost:8080/manifest.json\n")
		panic(http.ListenAndServe(":8080", nil))

	case apps.DeployAWSLambda:
		lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
	}
}
