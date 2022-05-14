package awsapp

import (
	"errors"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

func RunAWSLambda(app *goapp.App) error {
	if app.Manifest.Deploy.AWSLambda == nil {
		return errors.New("no aws_lambda in the app's manifest")
	}

	app.Mode = apps.DeployAWSLambda
	lambda.Start(httpadapter.New(app.Router).Proxy)
	return nil
}
