// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httpout

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/services/httpservice"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
)

type Service interface {
	config.Configurable
	httpservice.HTTPService

	GetFromURL(url string, trusted bool, limit int64) ([]byte, error)
}

type service struct {
	httpservice.HTTPService

	conf config.Service
}

var _ config.Configurable = (*service)(nil)
var _ httpservice.HTTPService = (*service)(nil)

func NewService(conf config.Service) Service {
	return &service{
		HTTPService: httpservice.MakeHTTPService(conf.MattermostConfig()),
		conf:        conf,
	}
}

func (s *service) Configure(_ config.Config) error {
	s.HTTPService = httpservice.MakeHTTPService(s.conf.MattermostConfig())
	return nil
}

func (s *service) GetFromURL(url string, trusted bool, limit int64) ([]byte, error) {
	client := s.MakeClient(trusted)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errData, _ := httputils.LimitReadAll(resp.Body, limit)
		return nil, errors.Errorf("received status %v: %v", resp.Status, string(errData))
	}

	return httputils.LimitReadAll(resp.Body, limit)
}
