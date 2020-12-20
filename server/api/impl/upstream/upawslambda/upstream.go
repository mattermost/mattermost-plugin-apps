// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/upstream"
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

func (u *Upstream) Notify(call *api.Call) error {
	_, err := u.invoke(call, true)
	return err
}

func (u *Upstream) Call(call *api.Call) *api.CallResponse {
	bb, err := u.invoke(call, false)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	cr := api.CallResponse{}
	err = json.Unmarshal(bb, &cr)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	return &cr
}

func (u *Upstream) GetBindings(call *api.Call) ([]*api.Binding, error) {
	bb, err := u.invoke(call, false)
	if err != nil {
		return nil, err
	}
	return upstream.DecodeBindingsResponse(bytes.NewReader(bb))
}

func (u *Upstream) invoke(call *api.Call, asNotification bool) ([]byte, error) {
	funcName := call.URL
	invocationType := lambda.InvocationTypeRequestResponse
	if asNotification {
		invocationType = lambda.InvocationTypeEvent
	}
	return u.aws.InvokeLambda(string(u.app.Manifest.AppID), funcName, invocationType, call)
}
