// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

type Proxy interface {
	GetBindings(*Context) ([]*Binding, error)
	Call(SessionToken, *Call) *CallResponse
	Notify(cc *Context, subj Subject) error
	Asset(AppID, string) ([]byte, error)

	ProvisionBuiltIn(AppID, Upstream)
}
