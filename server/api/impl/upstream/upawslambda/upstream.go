// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

type Upstream struct {
	app       *apps.App
	awsClient awsclient.Client
}

// invocationPayload is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyRequest
type invocationPayload struct {
	Path       string            `json:"path"`
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       interface{}       `json:"body"`
}

// invocationResponse is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyResponse
type invocationResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func NewUpstream(app *apps.App, awsClient awsclient.Client) *Upstream {
	return &Upstream{
		app:       app,
		awsClient: awsClient,
	}
}

func (u *Upstream) OneWay(call *apps.Call) error {
	payload, err := callToInvocationPayload(call)
	if err != nil {
		return errors.Wrap(err, "failed to covert call into invocation payload")
	}

	name, err := awsclient.MakeLambdaName(u.app.AppID, u.app.Version, u.app.Functions[0].Name)
	if err != nil {
		return err
	}
	_, err = u.awsClient.InvokeLambda(name, lambda.InvocationTypeEvent, payload)
	return err
}

func (u *Upstream) Roundtrip(call *apps.Call) (io.ReadCloser, error) {
	payload, err := callToInvocationPayload(call)
	if err != nil {
		return nil, errors.Wrap(err, "failed to covert call into invocation payload")
	}

	name, err := awsclient.MakeLambdaName(u.app.AppID, u.app.Version, u.app.Functions[0].Name)
	if err != nil {
		return nil, err
	}
	bb, err := u.awsClient.InvokeLambda(name, lambda.InvocationTypeRequestResponse, payload)
	if err != nil {
		return nil, err
	}

	var resp invocationResponse
	err = json.Unmarshal(bb, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "Error marshaling request payload")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("lambda invocation failed with status code %v and body %v", resp.StatusCode, resp.Body)
	}

	return ioutil.NopCloser(strings.NewReader(resp.Body)), nil
}

func callToInvocationPayload(call *apps.Call) ([]byte, error) {
	request := invocationPayload{
		Path:       call.Path,
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       call,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal call into http payload")
	}

	return payload, nil
}
