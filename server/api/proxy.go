// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

import (
	"io"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

type Proxy interface {
	AddBuiltinUpstream(apps.AppID, Upstream)
	Call(apps.SessionToken, *apps.CallRequest) *apps.CallResponse
	GetAsset(apps.AppID, string) (io.ReadCloser, int, error)
	GetBindings(apps.SessionToken, *apps.Context) ([]*apps.Binding, error)
	Notify(cc *apps.Context, subj apps.Subject) error
}
