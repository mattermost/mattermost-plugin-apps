package config

import (
	"path/filepath"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/i18n"

	"github.com/mattermost/mattermost/server/v8/platform/services/configservice"

	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type TestService struct {
	config    Config
	i18n      *i18n.Bundle
	log       utils.Logger
	mm        *pluginapi.Client
	mmconfig  model.Config
	telemetry *telemetry.Telemetry
}

var _ Service = (*TestService)(nil)

func NewTestConfigService(testConfig *Config) *TestService {
	conf, _ := NewTestService(testConfig)
	return conf
}

func NewTestService(testConfig *Config) (*TestService, *plugintest.API) {
	testAPI := &plugintest.API{}
	testDriver := &plugintest.Driver{}
	if testConfig == nil {
		testConfig = &Config{}
	}

	testAPI.On("GetBundlePath").Return("/", nil)
	i18nBundle, _ := i18n.InitBundle(testAPI, filepath.Join("assets", "i18n"))

	return &TestService{
		config:    *testConfig,
		i18n:      i18nBundle,
		log:       utils.NewTestLogger(),
		mm:        pluginapi.NewClient(testAPI, testDriver),
		telemetry: telemetry.NewTelemetry(nil),
	}, testAPI
}

func (s TestService) WithMattermostConfig(mmconfig model.Config) *TestService {
	s.mmconfig = mmconfig
	return &s
}

func (s TestService) WithMattermostAPI(mm *pluginapi.Client) *TestService {
	s.mm = mm
	return &s
}

func (s *TestService) Get() Config {
	return s.config
}

func (s *TestService) NewBaseLogger() utils.Logger {
	return s.log
}

func (s *TestService) MattermostAPI() *pluginapi.Client {
	return s.mm
}

func (s *TestService) I18N() *i18n.Bundle {
	return s.i18n
}

func (s *TestService) Telemetry() *telemetry.Telemetry {
	return s.telemetry
}

func (s *TestService) MattermostConfig() configservice.ConfigService {
	return &mattermostConfigService{&s.mmconfig}
}

func (s *TestService) Reconfigure(StoredConfig, bool, ...Configurable) error {
	return nil
}

func (s *TestService) StoreConfig(sc StoredConfig, _ utils.Logger) error {
	s.config.StoredConfig = sc
	return nil
}

func (s *TestService) SystemDefaultFlags() (bool, bool) { return false, false }
