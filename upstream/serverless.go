// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// invocationPayload is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyRequest
type ServerlessRequest struct {
	Path       string            `json:"path"`
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// invocationResponse is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyResponse
type ServerlessResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	Body            string            `json:"body"`
}

func ServerlessResponseFromJSON(data []byte) (*ServerlessResponse, error) {
	resp := ServerlessResponse{}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding serverless response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("function failed with status code %v, body %v", resp.StatusCode, resp.Body)
	}
	return &resp, nil
}

func ServerlessRequestFromCall(creq apps.CallRequest) ([]byte, error) {
	body, err := json.Marshal(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode serverless request")
	}
	request := ServerlessRequest{
		Path:       creq.Path,
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode serverless request")
	}
	return payload, nil
}
