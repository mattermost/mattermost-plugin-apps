package configurator

import "github.com/mattermost/mattermost-server/v5/model"

type testConfigurator struct {
	config *Config
}

var _ Service = (*testConfigurator)(nil)

func NewTestConfigurator(config *Config) Service {
	return &testConfigurator{
		config: config,
	}
}

func (c *testConfigurator) GetConfig() Config {
	return *c.config
}

func (c *testConfigurator) GetMattermostConfig() *model.Config {
	return &model.Config{}
}

func (c *testConfigurator) Refresh() error {
	return nil
}

func (c *testConfigurator) Store(newStored Mapper) error {
	return nil
}
