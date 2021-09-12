// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// DeployType determines how Apps are deployed and accessed.
type DeployType string

const (
	// HTTP-deployable app. All communications are done via HTTP requests. Paths
	// for both functions and static assets are appended to RootURL "as is".
	// Mattermost authenticates to the App with an optional shared secret based
	// JWT.
	DeployHTTP DeployType = "http"

	// AWS Lambda-deployable app. All functions are called via AWS Lambda
	// "Invoke" API, using path mapping provided in the app's manifest. Static
	// assets are served out of AWS S3, using the "Download" method. Mattermost
	// authenticates to AWS, no authentication to the App is necessary.
	DeployAWSLambda DeployType = "aws_lambda"

	// Kubeless-deployable app.
	DeployKubeless DeployType = "kubeless"

	// Builtin app. All functions and resources are served by directly invoking
	// go functions. No manifest, no Mattermost to App authentication are
	// needed.
	DeployBuiltin DeployType = "builtin"

	// An App running as a plugin. All communications are done via inter-plugin HTTP requests.
	// Authentication is done via the plugin.Context.SourcePluginId field.
	DeployPlugin DeployType = "plugin"
)

// Deploy contains App's deployment data, only the fields supported by the App
// should be populated.
type Deploy struct {
	// HTTP contains metadata for an app that is already, deployed externally
	// and us accessed over HTTP. The JSON name `http` must match the type.
	HTTP *HTTP `json:"http,omitempty"`

	// AWSLambda contains metadata for an app that can be deployed to AWS Lambda
	// and S3 services, and is accessed using the AWS APIs. The JSON name
	// `aws_lambda` must match the type.
	AWSLambda *AWSLambda `json:"aws_lambda,omitempty"`

	// Kubeless contains metadata for an app that can be deployed to Kubeless
	// running on a Kubernetes cluster, and is accessed using the Kubernetes
	// APIs and HTTP. The JSON name `kubeless` must match the type.
	Kubeless *Kubeless `json:"kubeless,omitempty"`

	// Plugin contains metadata for an app that is implemented and is deployed
	// and accessed as a local Plugin. The JSON name `plugin` must match the
	// type.
	Plugin *Plugin `json:"plugin,omitempty"`
}

func (t DeployType) Validate() error {
	switch t {
	case DeployHTTP,
		DeployAWSLambda,
		DeployBuiltin,
		DeployKubeless,
		DeployPlugin:
		return nil
	default:
		return utils.NewInvalidError("%s is not a valid app type", t)
	}
}

func (t DeployType) String() string {
	switch t {
	case DeployHTTP:
		return "HTTP"
	case DeployAWSLambda:
		return "AWS Lambda"
	case DeployBuiltin:
		return "Built-in"
	case DeployKubeless:
		return "Kubeless"
	case DeployPlugin:
		return "Mattermost Plugin"
	default:
		return string(t)
	}
}
