package configurator

import (
	"github.com/mattermost/mattermost-plugin-apps/server/apps"
	"github.com/mattermost/mattermost-server/v5/model"
)

type testConfigurator struct {
	config *apps.Config
}

var _ apps.Configurator = (*testConfigurator)(nil)

func NewTestConfigurator(config *apps.Config) apps.Configurator {
	return &testConfigurator{
		config: config,
	}
}

func (c *testConfigurator) GetConfig() apps.Config {
	return *c.config
}

func (c *testConfigurator) GetMattermostConfig() *model.Config {
	return &model.Config{}
}

func (c *testConfigurator) RefreshConfig(*apps.StoredConfig) error {
	return nil
}

func (c *testConfigurator) StoreConfig(newStored apps.ConfigMapper) error {
	return nil
}
