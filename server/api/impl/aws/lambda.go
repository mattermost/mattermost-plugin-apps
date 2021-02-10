// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// InvokeLambda runs a lambda function with specified name and returns a payload
func (c *Client) InvokeLambda(appID apps.AppID, appVersion apps.AppVersion, functionName, invocationType string, request interface{}) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling request payload")
	}

	name, err := getFunctionName(appID, appVersion, functionName)
	if err != nil {
		return nil, errors.Wrap(err, "can't get function name")
	}

	result, err := c.Service().lambda.Invoke(&lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: aws.String(invocationType),
		Payload:        payload,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Error calling function %s", name)
	}
	return result.Payload, nil
}
