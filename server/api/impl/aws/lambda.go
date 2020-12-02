// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// InvokeFunction runs a lambda function with specified name and returns a payload
func (s *Service) InvokeLambda(functionName string, request interface{}) ([]byte, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling request payload")
	}

	result, err := s.lambda().Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: payload})
	if err != nil {
		return nil, errors.Wrapf(err, "Error calling function %s", functionName)
	}
	return result.Payload, nil
}
