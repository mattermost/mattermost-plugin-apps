// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"encoding/json"
	"errors"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

func Notify(u api.Upstream, call *api.Call) error {
	return u.OneWay(call)
}

func Call(u api.Upstream, call *api.Call) *api.CallResponse {
	r, err := u.Roundtrip(call)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	defer r.Close()

	cr := api.CallResponse{}
	err = json.NewDecoder(r).Decode(&cr)
	if err != nil {
		return api.NewErrorCallResponse(err)
	}
	return &cr
}

func GetBindings(u api.Upstream, call *api.Call) ([]*api.Binding, error) {
	r, err := u.Roundtrip(call)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	cr := api.CallResponse{
		Data: &[]*api.Binding{},
	}
	err = json.NewDecoder(r).Decode(&cr)
	if err != nil {
		return nil, err
	}

	bindings, ok := cr.Data.(*[]*api.Binding)
	if !ok {
		return nil, errors.New("failed to decode bindings")
	}
	return *bindings, nil
}
