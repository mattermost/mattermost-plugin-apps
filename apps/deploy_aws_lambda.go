// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// AWSLambda contains metadata for an app that can be deployed to AWS Lambda
// and S3 services, and is accessed using the AWS APIs. The JSON name
// `aws_lambda` must match the type.
type AWSLambda struct {
	Functions []AWSLambdaFunction `json:"functions,omitempty"`
}

func (a *AWSLambda) Validate() error {
	if a == nil {
		return nil
	}
	var result error
	if len(a.Functions) == 0 {
		result = multierror.Append(result,
			utils.NewInvalidError("must provide at least 1 function in aws_lambda.Functions"))
	}
	for _, f := range a.Functions {
		err := f.Validate()
		if err != nil {
			result = multierror.Append(result,
				errors.Wrapf(err, "%q is not valid", f.Name))
		}
	}
	return result
}

// AWSLambdaFunction describes a distinct AWS Lambda function defined by the
// app, and what path should be mapped to it. See
// https://developers.mattermost.com/integrate/apps/deployment/#making-your-app-runnable-as-an-aws-lambda-function
// for more information.
//
// cmd/appsctl will create or update the manifest's aws_lambda functions in the
// AWS Lambda service.
//
// upawslambda will use the manifest's aws_lambda functions to find the closest
// match for the call's path, and then to invoke the AWS Lambda function.
type AWSLambdaFunction struct {
	// The lambda function with its Path the longest-matching prefix of the
	// call's Path will be invoked for a call.
	Path    string `json:"path"`
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

func (f AWSLambdaFunction) Validate() error {
	var result error
	if f.Path == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("aws_lambda path must not be empty"))
	}
	if f.Name == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("aws_lambda name must not be empty"))
	}
	if f.Handler == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("aws_lambda handler must not be empty"))
	}
	if f.Runtime == "" {
		result = multierror.Append(result,
			utils.NewInvalidError("aws_lambda runtime must not be empty"))
	}
	return result
}
