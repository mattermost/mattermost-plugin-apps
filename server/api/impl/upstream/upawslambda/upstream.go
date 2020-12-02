// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upawslambda

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type Upstream struct{}

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) (*api.CallResponse, error) {
	cr := api.CallResponse{}
	// err := json.NewDecoder(resp.Body).Decode(&cr)
	// if err != nil {
	// 	return nil, err
	// }
	return &cr, nil
}

func (u *Upstream) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	out := []*api.Binding{}
	// err := json.NewDecoder(resp.Body).Decode(&out)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "error unmarshalling function")
	// }
	return out, nil
}
