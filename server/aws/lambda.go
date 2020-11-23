// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// CreateFunction method creates lambda function
func (a *Client) CreateFunction(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" || handler == "" || resource == "" || runtime == "" {
		return errors.Errorf("you must supply a zip file, function name, handler, ARN and runtime - %s %s %s %s %s", zipFile, function, handler, resource, runtime)
	}

	s := ""
	s += fmt.Sprintf("func = %s\nhandler = %s\nresource = %s\nruntime = %s\n", function, handler, resource, runtime)

	contents, err := ioutil.ReadAll(zipFile)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}
	s += "content = " + strconv.Itoa(len(contents)) + "\n"

	createCode := &lambda.FunctionCode{
		// S3ObjectVersion: aws.String(""), //TODO
		ZipFile: contents,
	}

	createArgs := &lambda.CreateFunctionInput{
		Code:         createCode,
		FunctionName: &function,
		Handler:      &handler,
		Role:         &resource,
		Runtime:      &runtime,
	}

	result, err := a.Service().lambda.CreateFunction(createArgs)
	if err != nil {
		return errors.Wrapf(err, "Can't create function res = %v\n additional = \n%v", result, s)
		// return errors.Wrapf(err, "Can't create function additional")
	}
	a.logger.Infof("function named %s was created with result - %v", function, result)

	return nil
}

// InvokeFunction runs a lambda function with specified name and returns a payload
func (a *Client) InvokeFunction(functionName string, request interface{}) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshalling request payload")
	}

	result, err := a.Service().lambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		return nil, errors.Wrapf(err, "Error calling function %s", functionName)
	}
	return result.Payload, nil
}
