// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
)

type Upstream struct {
	app *api.App
	aws *aws.Client
}

func NewUpstream(app *api.App, aws *aws.Client) *Upstream {
	return &Upstream{
		app: app,
		aws: aws,
	}
}

func (u *Upstream) OneWay(call *api.Call) error {
	_, err := u.aws.InvokeLambda(string(u.app.Manifest.AppID), u.app.Manifest.Version, call.URL, lambda.InvocationTypeEvent, call)
	return err
}

func (u *Upstream) Roundtrip(call *api.Call) (io.ReadCloser, error) {
	bb, err := u.aws.InvokeLambda(string(u.app.Manifest.AppID), u.app.Manifest.Version, call.URL, lambda.InvocationTypeRequestResponse, call)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(bytes.NewReader(bb)), err
}
