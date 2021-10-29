// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
)

// HTTPCallRequest is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyRequest
type HTTPCallRequest struct {
	Path       string            `json:"path"`
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	RawQuery   string            `json:"rawQuery,omitempty"`
	Body       string            `json:"body"`
}

// HTTPCallResponse is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyResponse
type HTTPCallResponse struct {
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	Body            string            `json:"body"`
}

func HTTPCallResponseFromJSON(data []byte) (*HTTPCallResponse, error) {
	resp := HTTPCallResponse{}
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding JSON-encoded HTTP response")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("function failed with status code %v, body %v", resp.StatusCode, resp.Body)
	}
	return &resp, nil
}

func (creq CallRequest) ToHTTPCallRequestJSON() ([]byte, error) {
	body, err := json.Marshal(creq)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode HTTP request as JSON")
	}
	request := HTTPCallRequest{
		Path:       creq.Path,
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode HTTP call request as JSON")
	}
	return payload, nil
}
