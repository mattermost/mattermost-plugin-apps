// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package aws

import (
	"sync"

	"github.com/mattermost/mattermost-plugin-apps/awsclient"
)

type Service interface {
	Refresh() error
	Client() awsclient.Client
}

type service struct {
	awsAccessKeyID     string
	awsSecretAccessKey string
	logger             awsclient.Logger

	mutex  sync.RWMutex
	client awsclient.Client
}

func MakeService(awsAccessKeyID, awsSecretAccessKey string, logger awsclient.Logger) (Service, error) {
	s := &service{
		awsAccessKeyID:     awsAccessKeyID,
		awsSecretAccessKey: awsSecretAccessKey,
		logger:             logger,
	}
	err := s.Refresh()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *service) Refresh() error {
	client, err := awsclient.MakeClient(s.awsAccessKeyID, s.awsSecretAccessKey, s.logger)
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
