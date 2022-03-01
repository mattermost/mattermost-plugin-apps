// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/hashicorp/go-multierror"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// DeployType determines how Apps are deployed and accessed.
type DeployType string

type DeployTypes []DeployType

const (
	// AWS Lambda-deployable app. All functions are called via AWS Lambda
	// "Invoke" API, using path mapping provided in the app's manifest. Static
	// assets are served out of AWS S3, using the "Download" method. Mattermost
	// authenticates to AWS, no authentication to the App is necessary.
	DeployAWSLambda DeployType = "aws_lambda"

	// Builtin app. All functions and resources are served by directly invoking
	// go functions. No manifest, no Mattermost to App authentication are
	// needed.
	DeployBuiltin DeployType = "builtin"

	// HTTP-deployable app. All communications are done via HTTP requests. Paths
	// for both functions and static assets are appended to RootURL "as is".
	// Mattermost authenticates to the App with an optional shared secret based
	// JWT.
	DeployHTTP DeployType = "http"

	// OpenFaaS-deployable app.
	DeployOpenFAAS DeployType = "open_faas"

	// An App running as a plugin. All communications are done via inter-plugin HTTP requests.
	// Authentication is done via the plugin.Context.SourcePluginId field.
	DeployPlugin DeployType = "plugin"
)

var KnownDeployTypes = DeployTypes{
	DeployAWSLambda,
	DeployBuiltin,
	DeployHTTP,
	DeployOpenFAAS,
	DeployPlugin,
}

// Deploy contains App's deployment data, only the fields supported by the App
// should be populated.
type Deploy struct {
	// AWSLambda contains metadata for an app that can be deployed to AWS Lambda
	// and S3 services, and is accessed using the AWS APIs. The JSON name
	// `aws_lambda` must match the type.
	AWSLambda *AWSLambda `json:"aws_lambda,omitempty"`

	// HTTP contains metadata for an app that is already, deployed externally
	// and us accessed over HTTP. The JSON name `http` must match the type.
	HTTP *HTTP `json:"http,omitempty"`

	OpenFAAS *OpenFAAS `json:"open_faas,omitempty"`

	// Plugin contains metadata for an app that is implemented and is deployed
	// and accessed as a local Plugin. The JSON name `plugin` must match the
	// type.
	Plugin *Plugin `json:"plugin,omitempty"`
}

func (t DeployType) Validate() error {
	switch t {
	case DeployAWSLambda,
		DeployBuiltin,
		DeployHTTP,
		DeployOpenFAAS,
		DeployPlugin:
		return nil
	default:
		return utils.NewInvalidError("%s is not a valid app type", t)
	}
}

func (t DeployType) String() string {
	switch t {
	case DeployAWSLambda:
		return "AWS Lambda"
	case DeployBuiltin:
		return "Built-in"
	case DeployHTTP:
		return "HTTP"
	case DeployOpenFAAS:
		return "OpenFaaS"
	case DeployPlugin:
		return "Mattermost Plugin"
	default:
		return string(t)
	}
}

func (t DeployTypes) Contains(typ DeployType) bool {
	for _, current := range t {
		if current == typ {
			return true
		}
	}
	return false
}

func (d Deploy) Validate() error {
	var result error

	if d.AWSLambda == nil &&
		d.HTTP == nil &&
		d.OpenFAAS == nil &&
		d.Plugin == nil {
		result = multierror.Append(result,
			utils.NewInvalidError("manifest has no deployment information (http, aws_lambda, open_faas, etc.)"))
	}

	for _, v := range []validator{
		d.AWSLambda,
		d.HTTP,
		d.OpenFAAS,
		d.Plugin,
	} {
		// Validate must ignore nil pointer in its implementation, v is never
		// nil (interface wrapper).
		if err := v.Validate(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result
}

func (d Deploy) MustDeployAs() DeployType {
	all := d.DeployTypes()
	if len(all) == 1 {
		return all[0]
	}
	return ""
}

func (d Deploy) DeployTypes() (out []DeployType) {
	if d.AWSLambda != nil {
		out = append(out, DeployAWSLambda)
	}
	if d.HTTP != nil {
		out = append(out, DeployHTTP)
	}
	if d.OpenFAAS != nil {
		out = append(out, DeployOpenFAAS)
	}
	if d.Plugin != nil {
		out = append(out, DeployPlugin)
	}
	return out
}

func (d Deploy) Contains(dtype DeployType) bool {
	switch dtype {
	case DeployAWSLambda:
		return d.AWSLambda != nil
	case DeployHTTP:
		return d.HTTP != nil
	case DeployOpenFAAS:
		return d.OpenFAAS != nil
	case DeployPlugin:
		return d.Plugin != nil
	}
	return false
}

func (d *Deploy) CopyType(src Deploy, typ DeployType) {
	switch typ {
	case DeployAWSLambda:
		d.AWSLambda = src.AWSLambda
	case DeployHTTP:
		d.HTTP = src.HTTP
	case DeployOpenFAAS:
		d.OpenFAAS = src.OpenFAAS
	case DeployPlugin:
		d.Plugin = src.Plugin
	}
}
