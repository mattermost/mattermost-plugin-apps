// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type Upstream struct{}

func NewUpstream(app *api.App) *Upstream {
	return &Upstream{}
}

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) *api.CallResponse {
	cr := api.CallResponse{}
	// err := json.NewDecoder(resp.Body).Decode(&cr)
	// if err != nil {
	// 	return nil, err
	// }
	return &cr
}

func (u *Upstream) GetBindings(call *api.Call) ([]*api.Binding, error) {
	return []*api.Binding{}, nil
}
