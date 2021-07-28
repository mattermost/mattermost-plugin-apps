package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// AppType is the type of an app: http, aws_lambda, or builtin.
type AppType string

const (
	// HTTP app (default). All communications are done via HTTP requests. Paths
	// for both functions and static assets are appended to RootURL "as is".
	// Mattermost authenticates to the App with an optional shared secret based
	// JWT.
	AppTypeHTTP AppType = "http"

	// AWS Lambda app. All functions are called via AWS Lambda "Invoke" API,
	// using path mapping provided in the app's manifest. Static assets are
	// served out of AWS S3, using the "Download" method. Mattermost
	// authenticates to AWS, no authentication to the App is necessary.
	AppTypeAWSLambda AppType = "aws_lambda"

	AppTypeKubeless AppType = "kubeless"

	// Builtin app. All functions and resources are served by directly invoking
	// go functions. No manifest, no Mattermost to App authentication are
	// needed.
	AppTypeBuiltin AppType = "builtin"

	// An App running as a plugin. All communications are done via inter-plugin HTTP requests.
	// Authentication is done via the plugin.Context.SourcePluginId field.
	AppTypePlugin AppType = "plugin"
)

func (at AppType) IsValid() error {
	switch at {
	case AppTypeHTTP, AppTypeAWSLambda, AppTypeBuiltin, AppTypeKubeless, AppTypePlugin:
		return nil
	default:
		return utils.NewInvalidError("%s is not a valid app type", at)
	}
}
