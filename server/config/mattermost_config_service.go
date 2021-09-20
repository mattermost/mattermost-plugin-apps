// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package config

import (
	"crypto/ecdsa"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/services/configservice"
)

type mattermostConfigService struct {
	mmconf *model.Config
}

var _ configservice.ConfigService = (*mattermostConfigService)(nil)

func (c *mattermostConfigService) Config() *model.Config {
	return c.mmconf
}

func (c *mattermostConfigService) AddConfigListener(func(old, current *model.Config)) string {
	return ""
}

func (c *mattermostConfigService) RemoveConfigListener(string) {}
func (c *mattermostConfigService) AsymmetricSigningKey() *ecdsa.PrivateKey {
	return nil
}
