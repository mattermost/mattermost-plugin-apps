// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/api/impl/aws"
)

type Upstream struct {
	app *api.App
	aws *aws.Service
}

func NewUpstream(app *api.App, aws *aws.Service) *Upstream {
	return &Upstream{
		app: app,
		aws: aws,
	}
}

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	// Assuming function name is n.Subject
	funcName := string(n.Subject)
	if _, err := u.aws.InvokeLambda(string(u.app.Manifest.AppID), funcName, lambda.InvocationTypeEvent, n); err != nil {
		return api.NewErrorCallResponse(err)
	}
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) *api.CallResponse {
	cr := api.CallResponse{}
	// Assuming that in case of lambda invocation URL will have the lambda function name.
	// We probably should change name URL to something more clear
	// Or we could add another field in the call struct
	funcName := call.URL
	resp, err := u.aws.InvokeLambda(string(u.app.Manifest.AppID), funcName, lambda.InvocationTypeRequestResponse, call)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}

	err = json.Unmarshal(resp, &cr)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	return &cr
}

func (u *Upstream) GetBindings(call *api.Call) ([]*api.Binding, error) {
	return []*api.Binding{}, nil
}
