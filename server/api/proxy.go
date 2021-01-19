// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/modelapps"

type Proxy interface {
	GetBindings(*modelapps.Context) ([]*modelapps.Binding, error)
	Call(modelapps.SessionToken, *modelapps.Call) *modelapps.CallResponse
	Notify(cc *modelapps.Context, subj modelapps.Subject) error

	ProvisionBuiltIn(modelapps.AppID, Upstream)
}
