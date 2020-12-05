// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/pkg/errors"
)

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	// TODO
	resp, err := u.post("", u.rootURL+"/notify/"+string(n.Subject), n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) *api.CallResponse {
	resp, err := u.post(call.Context.ActingUserID, u.rootURL+call.URL, call)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	defer resp.Body.Close()

	cr := api.CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	return &cr
}

func (u *Upstream) GetBindings(call *api.Call) ([]*api.Binding, error) {
	resp, err := u.post(call.Context.ActingUserID, u.rootURL+call.URL, call)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cr := api.CallResponse{
		Data: []*api.Binding{},
	}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return nil, err
	}
	bindings, ok := cr.Data.([]*api.Binding)
	if !ok {
		return nil, errors.Errorf("failed to decode bindings for app %s", call.Context.AppID)
	}
	return bindings, nil
}
