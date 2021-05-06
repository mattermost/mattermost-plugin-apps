// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httpout

import (
	"io"

	"github.com/mattermost/mattermost-server/v5/services/httpservice"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

type Service interface {
	config.Configurable
	httpservice.HTTPService

	GetFromURL(url string, trusted bool) ([]byte, error)
}

type service struct {
	httpservice.HTTPService

	conf config.Service
}

var _ config.Configurable = (*service)(nil)
var _ httpservice.HTTPService = (*service)(nil)

func NewService(conf config.Service) Service {
	return &service{
		HTTPService: httpservice.MakeHTTPService(conf.GetMattermostConfig()),
		conf:        conf,
	}
}

func (s *service) Configure(x config.Config) {
	s.HTTPService = httpservice.MakeHTTPService(s.conf.GetMattermostConfig())
}

func (s *service) GetFromURL(url string, trusted bool) ([]byte, error) {
	client := s.MakeClient(trusted)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer _ = resp.Body.Close()
	return io.ReadAll(resp.Body)
}
