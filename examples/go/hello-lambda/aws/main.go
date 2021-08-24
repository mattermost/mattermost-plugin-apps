package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	_ "github.com/mattermost/mattermost-plugin-apps/examples/go/hello-lambda"
)

func main() {
	lambda.Start(httpadapter.New(http.DefaultServeMux).Proxy)
}
