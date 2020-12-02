// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package api

type Proxy interface {
	GetManifest(manifestURL string) (*Manifest, error)
	GetBindings(*Context) ([]*Binding, error)
	Call(*Call) (*CallResponse, error)
	Notify(cc *Context, subj Subject) error

	ProvisionBuiltIn(AppID, Upstream)
}
