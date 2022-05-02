package awsapp

import (
	"errors"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"go.uber.org/zap/zapcore"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func RunAWSLambda(app *goapp.App) error {
	if app.Deploy.AWSLambda == nil {
		return errors.New("no aws_lambda in the app's manifest")
	}

	app.Mode = apps.DeployAWSLambda
	app.Log = utils.MustMakeCommandLogger(zapcore.DebugLevel)
	lambda.Start(httpadapter.New(app.Router).Proxy)
	return nil
}
