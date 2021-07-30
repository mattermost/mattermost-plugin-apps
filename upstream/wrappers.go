// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package upstream

import (
	"encoding/json"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func Notify(u Upstream, app apps.App, call apps.CallRequest) error {
	r, err := u.Roundtrip(app, call, true)
	if r != nil {
		r.Close()
	}
	return err
}

func Call(u Upstream, app apps.App, call apps.CallRequest) apps.CallResponse {
	r, err := u.Roundtrip(app, call, false)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	defer r.Close()

	cr := apps.CallResponse{}
	err = json.NewDecoder(r).Decode(&cr)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	return cr
}
