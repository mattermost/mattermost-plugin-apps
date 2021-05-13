// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

// InvokeLambda runs a lambda function with specified name and returns a payload
func (c *client) InvokeLambda(name, invocationType string, payload []byte) ([]byte, error) {
	result, err := c.lambda.Invoke(&lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String(invocationType),
		Payload:        payload,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Error calling function %s", name)
	}
	return result.Payload, nil
}

// CreateLambda method creates lambda function
func (c *client) CreateLambda(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" || handler == "" || resource == "" || runtime == "" {
		return errors.Errorf("you must supply a zip file, function name, handler, ARN and runtime - %p %s %s %s %s", zipFile, function, handler, resource, runtime)
	}

	contents, err := io.ReadAll(zipFile)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}

	createCode := &lambda.FunctionCode{
		ZipFile: contents,
	}

	createArgs := &lambda.CreateFunctionInput{
		Code:         createCode,
		FunctionName: &function,
		Handler:      &handler,
		Role:         &resource,
		Runtime:      &runtime,
	}

	result, err := c.lambda.CreateFunction(createArgs)
	if err != nil {
		return errors.Wrapf(err, "can't create function, %+v\n", result)
	}
	c.logger.Info(fmt.Sprintf("created function %s", function))
	return nil
}

func (c *client) CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" {
		return errors.New("you must supply a zip file and the function name")
	}

	_, err := c.lambda.GetFunction(&lambda.GetFunctionInput{FunctionName: &function})
	if err != nil {
		if _, ok := err.(*lambda.ResourceNotFoundException); !ok {
			return errors.Wrap(err, "failed go get function")
		}
		return c.CreateLambda(zipFile, function, handler, runtime, resource)
	}

	contents, err := io.ReadAll(zipFile)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}
	_, err = c.lambda.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		ZipFile:      contents,
		FunctionName: &function,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update function %v", function)
	}
	c.logger.Info(fmt.Sprintf("updated function %s", function))
	return nil
}
