// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/upstream"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

// Upstream wraps an awsClient to make requests to the App. It should not be
// reused between requests, nor cached.
type Upstream struct {
	app       *apps.App
	awsClient aws.Client
	bucket    string
}

var _ upstream.Upstream = (*Upstream)(nil)

// invocationPayload is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyRequest
type invocationPayload struct {
	Path       string            `json:"path"`
	HTTPMethod string            `json:"httpMethod"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// invocationResponse is a scoped down version of
// https://pkg.go.dev/github.com/aws/aws-lambda-go@v1.13.3/events#APIGatewayProxyResponse
type invocationResponse struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func NewUpstream(app *apps.App, awsClient aws.Client, bucket string) *Upstream {
	return &Upstream{
		app:       app,
		awsClient: awsClient,
		bucket:    bucket,
	}
}

func (u *Upstream) Roundtrip(call *apps.CallRequest, async bool) (io.ReadCloser, error) {
	typ := lambda.InvocationTypeRequestResponse
	if async {
		typ = lambda.InvocationTypeEvent
	}
	name := match(call.Path, u.app)
	if name == "" {
		return nil, utils.ErrNotFound
	}

	payload, err := callToInvocationPayload(call)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert call into invocation payload")
	}

	bb, err := u.awsClient.InvokeLambda(name, typ, payload)
	if async || err != nil {
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

	return io.NopCloser(strings.NewReader(resp.Body)), nil
}

func (u *Upstream) GetStatic(path string) (io.ReadCloser, int, error) {
	key := apps.AssetS3Name(u.app.AppID, u.app.Version, path)
	data, err := u.awsClient.GetS3(u.bucket, key)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "can't download from S3:bucket:%s, path:%s", u.bucket, path)
	}
	return io.NopCloser(bytes.NewReader(data)), http.StatusOK, nil
}

func callToInvocationPayload(call *apps.CallRequest) ([]byte, error) {
	body, err := json.Marshal(call)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal call for lambda payload")
	}

	request := invocationPayload{
		Path:       call.Path,
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal lambda payload")
	}

	return payload, nil
}

func match(callPath string, app *apps.App) string {
	matchedName := ""
	matchedPath := ""
	for _, f := range app.AWSLambda {
		if strings.HasPrefix(callPath, f.Path) {
			if len(f.Path) > len(matchedPath) {
				matchedName = apps.LambdaName(app.AppID, app.Version, f.Name)
				matchedPath = f.Path
			}
		}
	}
	return matchedName
}
