// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const MaxLambdaName = 64

// InvokeLambda runs a lambda function with specified name and returns a payload
func (c *client) InvokeLambda(ctx context.Context, name, invocationType string, payload []byte) ([]byte, error) {
	result, err := c.lambda.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String(invocationType),
		Payload:        payload,
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == lambda.ErrCodeResourceNotFoundException {
			return nil, utils.NewNotFoundError(awsErr)
		}
		return nil, errors.Wrapf(err, "invoke AWS Lambda function %s", name)
	}
	return result.Payload, nil
}

// CreateLambda method creates lambda function
func (c *client) CreateLambda(archive io.Reader, name, handler, runtime string, roleARN ARN) (ARN, error) {
	if archive == nil || name == "" || handler == "" || roleARN == "" || runtime == "" {
		return "", errors.Errorf("you must supply an archive (.zip or .jar) file, function name, handler, role ARN and runtime - %p %q %q %q %q", archive, name, handler, roleARN, runtime)
	}

	contents, err := io.ReadAll(archive)
	if err != nil {
		return "", errors.Wrap(err, "could not read archive file")
	}

	createCode := &lambda.FunctionCode{
		ZipFile: contents,
	}

	createArgs := &lambda.CreateFunctionInput{
		Code:         createCode,
		FunctionName: aws.String(name),
		Handler:      aws.String(handler),
		Role:         roleARN.AWSString(),
		Runtime:      aws.String(runtime),
	}

	fc, err := c.lambda.CreateFunction(createArgs)
	if err != nil {
		return "", errors.Wrapf(err, "can't create function, %+v", fc)
	}
	c.log.Infow("created function", "ARN", *fc.FunctionArn)
	return ARN(*fc.FunctionArn), nil
}

func (c *client) CreateOrUpdateLambda(zipFile io.Reader, function, handler, runtime string, roleARN ARN) (ARN, error) {
	if zipFile == nil || function == "" {
		return "", errors.New("you must supply a zip file and the function name")
	}

	fc, err := c.lambda.GetFunction(&lambda.GetFunctionInput{FunctionName: &function})
	if err != nil {
		if _, ok := err.(*lambda.ResourceNotFoundException); !ok {
			return "", errors.Wrap(err, "failed go get function")
		}
		return c.CreateLambda(zipFile, function, handler, runtime, roleARN)
	}

	contents, err := io.ReadAll(zipFile)
	if err != nil {
		return "", errors.Wrap(err, "could not read zip file")
	}
	_, err = c.lambda.UpdateFunctionCode(&lambda.UpdateFunctionCodeInput{
		ZipFile:      contents,
		FunctionName: &function,
	})
	if err != nil {
		return "", errors.Wrapf(err, "failed to update function %v", function)
	}
	c.log.Infow("updated function", "ARN", *fc.Configuration.FunctionArn)
	return ARN(*fc.Configuration.FunctionArn), nil
}

const updateWaitTime = 2 * time.Minute

// SetLambdaEnvironmentVariables sets environment variables for a lambda
// function. It waits until an function code deployment succeeds before updating
// the configuration.
func (c *client) SetLambdaEnvironmentVariables(arn string, started time.Time, vars map[string]*string) error {
	deadline := started.Add(updateWaitTime)
	retry := 5 * time.Second

RETRY:
	for time.Now().Before(deadline) {
		fc, err := c.lambda.GetFunctionConfiguration(&lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(arn),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to get function configuration for %s", arn)
		}

		switch *fc.LastUpdateStatus {
		case "Successful":
			break RETRY
		case "Failed":
			return errors.New("can't set environment variables after a failed deployment")
		default:
			c.log.Infof("function deployment %s, will wait %v", *fc.LastUpdateStatus, retry)
			time.Sleep(retry)
			retry *= 2
		}
	}

	_, err := c.lambda.UpdateFunctionConfiguration(&lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(arn),
		Environment: &lambda.Environment{
			Variables: vars,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update function configuration for %s", arn)
	}

	c.log.Infof("set %v environment variables on %s", len(vars), arn)
	return nil
}

// LambdaName generates function name for a specific app, name can be 64
// characters long.
func LambdaName(appID apps.AppID, version apps.AppVersion, function string) string {
	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")
	sanitizedFunction := strings.ReplaceAll(function, " ", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, sanitizedFunction)
	if len(name) <= MaxLambdaName {
		return name
	}

	functionNameLength := MaxLambdaName - len(appID) - len(version) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s_%s_%s", appID, version, hashString)
	return name
}
