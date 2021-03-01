// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package awsclient

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const LambdaFunctionFileNameMaxSize = 64
const AppIDLengthLimit = 32
const VersionFormat = "v00.00.000"

// InvokeLambda runs a lambda function with specified name and returns a payload
func (c *client) InvokeLambda(name, invocationType string, request interface{}) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling request payload")
	}

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

	contents, err := ioutil.ReadAll(zipFile)
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
		if _, ok := err.(*lambda.ResourceConflictException); !ok {
			return errors.Wrapf(err, "Can't create function res = %v\n", result)
		}
	}
	c.logger.Info(fmt.Sprintf("function named %s was created with result - %v", function, result))

	return nil
}

func (c *client) CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime, resource string) error {
	if zipFile == nil || function == "" {
		return errors.New("you must supply a zip file, function name, handler, ARN and runtime")
	}

	exists := false
	_, err := c.lambda.GetFunction(&lambda.GetFunctionInput{FunctionName: &function})
	if _, ok := err.(*lambda.ResourceNotFoundException); ok {
		exists = true
	} else {
		return errors.Wrap(err, "Failed go get function")
	}
	if !exists {
		return c.CreateLambda(zipFile, function, handler, runtime, resource)
	}

	contents, err := ioutil.ReadAll(zipFile)
	if err != nil {
		return errors.Wrap(err, "could not read zip file")
	}
	c.logger.Info("Updating existing function", "name", function)
	result, err := c.lambda.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		ZipFile:      contents,
		FunctionName: &function,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update function %v", function)
	}
	c.logger.Info(fmt.Sprintf("function named %s was updated", function), "result", result.String())
	return nil
}

// getFunctionName generates function name for a specific app
// name can be 64 characters long.
func MakeLambdaName(appID apps.AppID, version apps.AppVersion, function string) (string, error) {
	if len(appID) > AppIDLengthLimit {
		return "", errors.Errorf("appID %s too long, should be %d bytes", appID, AppIDLengthLimit)
	}
	if len(version) > len(VersionFormat) {
		return "", errors.Errorf("version %s too long, should be in %s format", version, VersionFormat)
	}

	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, function)
	if len(name) <= LambdaFunctionFileNameMaxSize {
		return name, nil
	}
	functionNameLength := LambdaFunctionFileNameMaxSize - len(sanitizedAppID) - len(sanitizedVersion) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s-%s-%s", sanitizedAppID, sanitizedVersion, hashString)
	return name, nil
}
