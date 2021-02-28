// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

const lambdaFunctionFileNameMaxSize = 64
const appIDLengthLimit = 32
const versionFormat = "v00.00.000"

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

	name, err := functionName(u.app.AppID, u.app.Version, u.app.Functions[0].Name)
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

	name, err := functionName(u.app.AppID, u.app.Version, u.app.Functions[0].Name)
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

func functionName(appID apps.AppID, version apps.AppVersion, function string) (string, error) {
	if len(appID) > appIDLengthLimit {
		return "", errors.Errorf("appID %s too long, should be %d bytes", appID, appIDLengthLimit)
	}
	if len(version) > len(versionFormat) {
		return "", errors.Errorf("version %s too long, should be in %s format", version, versionFormat)
	}

	// Sanitized any dots used in appID and version as lambda function names can not contain dots
	// While there are other non-valid characters, a dots is the most commonly used one
	sanitizedAppID := strings.ReplaceAll(string(appID), ".", "-")
	sanitizedVersion := strings.ReplaceAll(string(version), ".", "-")

	name := fmt.Sprintf("%s_%s_%s", sanitizedAppID, sanitizedVersion, function)
	if len(name) <= lambdaFunctionFileNameMaxSize {
		return name, nil
	}
	functionNameLength := lambdaFunctionFileNameMaxSize - len(sanitizedAppID) - len(sanitizedVersion) - 2
	hash := sha256.Sum256([]byte(name))
	hashString := hex.EncodeToString(hash[:])
	if len(hashString) > functionNameLength {
		hashString = hashString[:functionNameLength]
	}
	name = fmt.Sprintf("%s-%s-%s", sanitizedAppID, sanitizedVersion, hashString)
	return name, nil
}
