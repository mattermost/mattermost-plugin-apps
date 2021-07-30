// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upaws

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// Upstream wraps an awsClient to make requests to the App. It should not be
// reused between requests, nor cached.
type Upstream struct {
	awsClient      Client
	staticS3Bucket string
}

var _ upstream.Upstream = (*Upstream)(nil)

func MakeUpstream(accessKey, secret, region, staticS3bucket string, log utils.Logger) (*Upstream, error) {
	awsClient, err := MakeClient(accessKey, secret, region, log, "App Proxy")
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize AWS access")
	}
	return &Upstream{
		awsClient:      awsClient,
		staticS3Bucket: staticS3bucket,
	}, nil
}

func (u *Upstream) GetStatic(app apps.App, path string) (io.ReadCloser, int, error) {
	key := S3StaticName(app.AppID, app.Version, path)
	data, err := u.awsClient.GetS3(u.staticS3Bucket, key)
	if err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "can't download from S3:bucket:%s, path:%s", u.staticS3Bucket, path)
	}
	return io.NopCloser(bytes.NewReader(data)), http.StatusOK, nil
}

func (u *Upstream) Roundtrip(app apps.App, creq apps.CallRequest, async bool) (io.ReadCloser, error) {
	if app.Manifest.AWSLambda == nil {
		return nil, errors.New("no 'aws_lambda' section in manifest.json")
	}
	name := match(creq.Path, &app.Manifest)
	if name == "" {
		return nil, utils.ErrNotFound
	}

	data, err := u.invokeFunction(name, async, creq)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// InvokeFunction is a public method used in appsctl, but is not a part of the
// upstream.Upstream interface. It invokes a function with a specified name,
// with no conversion.
func (u *Upstream) invokeFunction(name string, async bool, creq apps.CallRequest) ([]byte, error) {
	typ := lambda.InvocationTypeRequestResponse
	if async {
		typ = lambda.InvocationTypeEvent
	}

	sreq, err := upstream.ServerlessRequestFromCall(creq)
	if err != nil {
		return nil, err
	}
	bb, err := u.awsClient.InvokeLambda(name, typ, sreq)
	if async || err != nil {
		return nil, err
	}
	resp, err := upstream.ServerlessResponseFromJSON(bb)
	if err != nil {
		return nil, err
	}
	return []byte(resp.Body), nil
}

func match(callPath string, m *apps.Manifest) string {
	matchedName := ""
	matchedPath := ""
	for _, f := range m.AWSLambda.Functions {
		if strings.HasPrefix(callPath, f.Path) {
			if len(f.Path) > len(matchedPath) {
				matchedName = LambdaName(m.AppID, m.Version, f.Name)
				matchedPath = f.Path
			}
		}
	}

	return matchedName
}
