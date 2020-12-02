// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package uphttp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func (u *Upstream) InvokeNotification(n *api.Notification) error {
	resp, err := u.post("", u.rootURL+"/notify/"+string(n.Subject), n)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (u *Upstream) InvokeCall(call *api.Call) (*api.CallResponse, error) {
	resp, err := u.post(call.Context.ActingUserID, u.rootURL+call.URL, call)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cr := api.CallResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

func (u *Upstream) GetBindings(cc *api.Context) ([]*api.Binding, error) {
	resp, err := u.get(cc.ActingUserID, appendGetContext(u.rootURL+api.AppBindingsPath, cc))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get bindings")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("returned with status %s", resp.Status)
	}

	out := []*api.Binding{}
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling function")
	}
	return out, nil
}
