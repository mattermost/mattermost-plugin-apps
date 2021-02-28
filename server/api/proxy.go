// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import "github.com/mattermost/mattermost-plugin-apps/apps"

type Proxy interface {
	GetBindings(*apps.Context) ([]*apps.Binding, error)
	Call(apps.SessionToken, *apps.Call) *apps.CallResponse
	Notify(cc *apps.Context, subj apps.Subject) error

	AddBuiltinUpstream(apps.AppID, Upstream)
}
