package apps

import (
	"github.com/pkg/errors"
)

// AWSLambdaFunction describes a distinct AWS Lambda function defined by the
// app, and what path should be mapped to it.
//
// cmd/appsctl will create or update the manifest's aws_lambda functions in the
// AWS Lambda service.
//
// upawslambda will use the manifest's aws_lambda functions to find the closest
// match for the call's path, and then to invoke the AWS Lambda function.
type AWSLambdaFunction struct {
	// The lambda function with its Path the longest-matching prefix of the
	// call's Path will be invoked for a call.
	Path string `json:"path"`

	// TODO @iomodo
	Name    string `json:"name"`
	Handler string `json:"handler"`
	Runtime string `json:"runtime"`
}

func (f AWSLambdaFunction) IsValid() error {
	if f.Path == "" {
		return errors.New("aws_lambda path must not be empty")
	}
	if f.Name == "" {
		return errors.New("aws_lambda name must not be empty")
	}
	if f.Runtime == "" {
		return errors.New("aws_lambda runtime must not be empty")
	}
	if f.Handler == "" {
		return errors.New("aws_lambda handler must not be empty")
	}
	return nil
}
