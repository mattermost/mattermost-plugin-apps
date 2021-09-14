// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/hashicorp/go-multierror"

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

	// OpenFaaS-deployable app.
	DeployOpenFAAS DeployType = "open_faas"

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

	OpenFAAS *OpenFAAS `json:"open_faas,omitempty"`

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
		DeployOpenFAAS,
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
	case DeployOpenFAAS:
		return "OpenFaaS"
	case DeployPlugin:
		return "Mattermost Plugin"
	default:
		return string(t)
	}
}

func (d Deploy) Validate() error {
	var result error

	if d.AWSLambda == nil &&
		d.HTTP == nil &&
		d.Kubeless == nil &&
		d.OpenFAAS == nil &&
		d.Plugin == nil {
		result = multierror.Append(result,
			utils.NewInvalidError("manifest has no deployment information (http, aws_lambda, open_faas, etc.)"))
	}

	for _, v := range []validator{
		d.AWSLambda,
		d.HTTP,
		d.Kubeless,
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
	if d.Kubeless != nil {
		out = append(out, DeployKubeless)
	}
	if d.OpenFAAS != nil {
		out = append(out, DeployOpenFAAS)
	}
	if d.Plugin != nil {
		out = append(out, DeployPlugin)
	}
	return out
}

func (d Deploy) SupportsDeploy(dtype DeployType) bool {
	switch dtype {
	case DeployAWSLambda:
		return d.AWSLambda != nil
	case DeployHTTP:
		return d.HTTP != nil
	case DeployKubeless:
		return d.Kubeless != nil
	case DeployOpenFAAS:
		return d.OpenFAAS != nil
	case DeployPlugin:
		return d.Plugin != nil
	}
	return false
}

func (d Deploy) UpdateDeploy(newDeploy Deploy, deployType DeployType) Deploy {
	var result Deploy
	if d.AWSLambda != nil || deployType == DeployAWSLambda {
		result.AWSLambda = newDeploy.AWSLambda
	}
	if d.HTTP != nil || deployType == DeployHTTP {
		result.HTTP = newDeploy.HTTP
	}
	if d.Kubeless != nil || deployType == DeployKubeless {
		result.Kubeless = newDeploy.Kubeless
	}
	if d.OpenFAAS != nil || deployType == DeployOpenFAAS {
		result.OpenFAAS = newDeploy.OpenFAAS
	}
	if d.Plugin != nil || deployType == DeployPlugin {
		result.Plugin = newDeploy.Plugin
	}
	return result
}
