package config

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type TestConfigurator struct {
	config   Config
	mmconfig model.Config
}

var _ Service = (*TestConfigurator)(nil)

func NewTestConfigurator(config Config) *TestConfigurator {
	return &TestConfigurator{
		config: config,
	}
}

func (c TestConfigurator) WithMattermostConfig(mmconfig model.Config) *TestConfigurator {
	c.mmconfig = mmconfig
	return &c
}

func (c *TestConfigurator) GetConfig() Config {
	return c.config
}

func (c *TestConfigurator) GetMattermostConfig() *model.Config {
	return &c.mmconfig
}

func (c *TestConfigurator) Reconfigure(StoredConfig, ...Configurable) error {
	return nil
}

func (c *TestConfigurator) StoreConfig(sc StoredConfig) error {
	c.config.StoredConfig = sc
	return nil
}
