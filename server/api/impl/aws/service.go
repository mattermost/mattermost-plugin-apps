// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"sync"

	"github.com/mattermost/mattermost-plugin-apps/awsclient"
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type Service interface {
	api.Configurable

	Client() awsclient.Client
}

type service struct {
	mutex  sync.RWMutex
	client awsclient.Client
	logger awsclient.Logger
}

func NewService(logger awsclient.Logger) Service {
	return &service{
		logger: logger,
	}
}

func (s *service) Configure(conf api.Config) error {
	client, err := awsclient.MakeClient(conf.AWSAccessKeyID, conf.AWSSecretAccessKey, s.logger)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	s.client = client
	s.mutex.Unlock()

	return nil
}

func (s *service) Client() awsclient.Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.client
}
