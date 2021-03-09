// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Proxy interface {
	Call(apps.SessionToken, *apps.Call) *apps.CallResponse
	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)
	GetBindings(*apps.Context) ([]*apps.Binding, error)
	Notify(cc *apps.Context, subj apps.Subject) error

	AddBuiltinUpstream(apps.AppID, Upstream)
}
