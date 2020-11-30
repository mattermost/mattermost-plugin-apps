// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	sdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/pkg/errors"
)

// DefaultRegion describes default region in aws
const DefaultRegion = "us-east-2"

type clientCache struct {
	store CredentialStoreService
	cache map[string]*Client // appID - Client
}

func newClientCache(store CredentialStoreService) (*clientCache, error) {
	cache := make(map[string]*Client)
	list, err := store.GetAppList()
	if err != nil {
		return nil, errors.Wrap(err, "can't get credentials")
	}
	cc := &clientCache{
		store: store,
		cache: cache,
	}
	for _, appID := range list {
		if _, err := cc.create(appID); err != nil {
			return nil, errors.Wrapf(err, "can't create client for app %s", appID)
		}
	}
	return cc, nil
}

func (c *clientCache) create(appID string) (*Client, error) {
	keyID, secret, err := c.store.GetAWSCredential(appID)
	if err != nil {
		return nil, errors.Wrap(err, "can't get app from the store")
	}
	var config *sdk.Config
	if keyID == "" || secret == "" {
		config = &sdk.Config{
			Region:      sdk.String(DefaultRegion),
			Credentials: credentials.NewEnvCredentials(), // Read Mattermost cloud credentials from the environment variables
		}
	} else {
		config = &sdk.Config{
			Region:      sdk.String(DefaultRegion),
			Credentials: credentials.NewStaticCredentials(keyID, secret, ""),
		}
	}
	client := NewAWSClientWithConfig(config)
	c.cache[appID] = client

	return client, nil
}

func (c *clientCache) get(appID string) (*Client, error) {
	client, ok := c.cache[appID]
	if !ok {
		return nil, errors.New("no client")
	}
	return client, nil
}
