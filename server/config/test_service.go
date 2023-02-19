package config

import (
	"path/filepath"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/services/configservice"

	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type TestService struct {
	api      API
	log      utils.Logger
	config   Config
	mmconfig model.Config
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
		config: *testConfig,
		api: API{
			I18N:       i18nBundle,
			Mattermost: pluginapi.NewClient(testAPI, testDriver),
			Telemetry:  telemetry.NewTelemetry(nil),
			Plugin:     testAPI,
		},
		log: utils.NewTestLogger(),
	}, testAPI
}

func (s TestService) WithMattermostConfig(mmconfig model.Config) *TestService {
	s.mmconfig = mmconfig
	return &s
}

func (s TestService) WithMattermostAPI(mm *pluginapi.Client) *TestService {
	s.api.Mattermost = mm
	return &s
}

func (s *TestService) GetMattermostConfig() configservice.ConfigService {
	return &mattermostConfigService{&s.mmconfig}
}

func (s *TestService) StoreConfig(sc StoredConfig, _ utils.Logger) error {
	s.config.StoredConfig = sc
	return nil
}

func (s *TestService) API() API                                              { return s.api }
func (s *TestService) Get() Config                                           { return s.config }
func (s *TestService) NewBaseLogger() utils.Logger                           { return s.log }
func (s *TestService) Reconfigure(StoredConfig, bool, ...Configurable) error { return nil }
func (s *TestService) SystemDefaultFlags() (bool, bool)                      { return false, false }
