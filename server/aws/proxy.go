// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/pkg/errors"
)

// Proxy is an interface to communicate with AWS lambda functions
type Proxy interface {
	IsAppInstalled(appID string) bool
	InstallApp(appID, releaseURL, keyID, secret string) error
	InvokeFunction(appID, functionName string, request interface{}) ([]byte, error)
}

type proxy struct {
	cc *clientCache
}

// NewAWSProxy creates proxy to communicate with AWS lambda functions
func NewAWSProxy(mm *pluginapi.Client) Proxy {
	store, err := newCredentialStoreService(mm)
	if err != nil {
		mm.Log.Error("can't create credential store", "err", err.Error())
		return nil
	}
	cc, err := newClientCache(store)
	if err != nil {
		mm.Log.Error("can't create client cache", "err", err.Error())
	}
	return &proxy{
		cc: cc,
	}
}

func (p *proxy) IsAppInstalled(appID string) bool {
	_, err := p.cc.get(appID)
	return err != nil
}

func (p *proxy) InstallApp(appID, releaseURL, keyID, secret string) error {
	if p.IsAppInstalled(appID) {
		return errors.New("app is already installed")
	}
	if err := p.cc.store.StoreAWSCredential(appID, keyID, secret); err != nil {
		return errors.Wrap(err, "can't store aws credentials in the KV store")
	}
	client, err := p.cc.create(appID)
	if err != nil {
		return errors.Wrapf(err, "can't create client for App %s and url %s", appID, releaseURL)
	}
	if err = client.InstallApp(releaseURL); err != nil {
		return errors.Wrap(err, "can't install App")
	}
	return nil
}

func (p *proxy) InvokeFunction(appID, functionName string, request interface{}) ([]byte, error) {
	if !p.IsAppInstalled(appID) {
		return nil, errors.New("app is not installed")
	}
	client, _ := p.cc.get(appID)
	name := createFunctionName(appID, functionName)
	return client.InvokeFunction(name, request)
}
