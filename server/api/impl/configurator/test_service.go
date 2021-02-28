package configurator

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type testConfigurator struct {
	config *api.Config
}

func NewTestConfigurator(config *api.Config) api.Configurator {
	return &testConfigurator{
		config: config,
	}
}

func (c *testConfigurator) GetConfig() api.Config {
	return *c.config
}

func (c *testConfigurator) GetMattermostConfig() *model.Config {
	return &model.Config{}
}

func (c *testConfigurator) Reconfigure(*api.StoredConfig, ...api.Configurable) error {
	return nil
}

func (c *testConfigurator) StoreConfig(sc *api.StoredConfig) error {
	c.config.StoredConfig = sc
	return nil
}
