package configurator

import (
	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

type testConfigurator struct {
	config *api.Config
}

var _ api.Configurator = (*testConfigurator)(nil)

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

func (c *testConfigurator) RefreshConfig(*api.StoredConfig) error {
	return nil
}

func (c *testConfigurator) StoreConfig(newStored api.ConfigMapper) error {
	return nil
}
