package config

import (
	"net"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/services/configservice"

	"github.com/mattermost/mattermost-plugin-apps/server/telemetry"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Configurable interface {
	Configure(Config, utils.Logger) error
}

type Service interface {
	Get() Config
	MattermostAPI() *pluginapi.Client
	MattermostConfig() configservice.ConfigService
	I18N() *i18n.Bundle
	Telemetry() *telemetry.Telemetry
	NewBaseLogger() utils.Logger
	SystemDefaultFlags() (devMode, allowHTTPApps bool)

	Reconfigure(_ StoredConfig, verbose bool, _ ...Configurable) error
	StoreConfig(StoredConfig, utils.Logger) error
}

var _ Service = (*service)(nil)

type service struct {
	pluginManifest model.Manifest
	botUserID      string
	mm             *pluginapi.Client
	i18n           *i18n.Bundle
	telemetry      *telemetry.Telemetry

	lock             *sync.RWMutex
	conf             *Config
	mattermostConfig *model.Config
}

func MakeService(mm *pluginapi.Client, pliginManifest model.Manifest, botUserID string, telemetry *telemetry.Telemetry, i18nBundle *i18n.Bundle, log utils.Logger) (Service, error) {
	s := &service{
		pluginManifest: pliginManifest,
		botUserID:      botUserID,
		mm:             mm,
		lock:           &sync.RWMutex{},
		i18n:           i18nBundle,
		telemetry:      telemetry,
	}

	sc := StoredConfig{}
	err := s.mm.Configuration.LoadPluginConfiguration(&sc)
	if err != nil {
		return nil, err
	}

	clone, err := s.newInitializedConfig(sc, log)
	if err != nil {
		return nil, err
	}

	s.lock.Lock()
	s.conf = clone
	s.lock.Unlock()

	return s, nil
}

func (s *service) newConfig() Config {
	return Config{
		PluginManifest: s.pluginManifest,
		BuildDate:      BuildDate,
		BuildHash:      BuildHash,
		BuildHashShort: BuildHashShort,
		BotUserID:      s.botUserID,
	}
}

func (s *service) newInitializedConfig(newStoredConfig StoredConfig, log utils.Logger) (*Config, error) {
	conf := s.newConfig()
	conf.StoredConfig = newStoredConfig
	newMattermostConfig := s.reloadMattermostConfig()

	if conf.DeveloperModeOverride != nil {
		conf.DeveloperMode = *conf.DeveloperModeOverride
	} else {
		conf.DeveloperMode = pluginapi.IsConfiguredForDevelopment(newMattermostConfig)
	}

	mattermostSiteURL := newMattermostConfig.ServiceSettings.SiteURL
	if mattermostSiteURL == nil {
		return nil, errors.New("plugin requires Mattermost Site URL to be set")
	}
	u, err := url.Parse(*mattermostSiteURL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid SiteURL in config")
	}

	var localURL string
	if newMattermostConfig.ServiceSettings.ConnectionSecurity != nil && *newMattermostConfig.ServiceSettings.ConnectionSecurity == model.ConnSecurityTLS {
		// If there is no reverse proxy use the server URL
		localURL = u.String()
	} else {
		// Avoid the reverse proxy by using the local port
		listenAddress := newMattermostConfig.ServiceSettings.ListenAddress
		if listenAddress == nil {
			return nil, errors.New("plugin requires Mattermost Listen Address to be set")
		}
		host, port, err := net.SplitHostPort(*listenAddress)
		if err != nil {
			return nil, err
		}

		if host == "" {
			host = "127.0.0.1"
		}

		localURL = "http://" + host + ":" + port + u.Path
	}

	conf.MattermostSiteURL = u.String()
	conf.MattermostLocalURL = localURL
	conf.PluginURLPath = "/plugins/" + conf.PluginManifest.Id
	conf.PluginURL = strings.TrimRight(u.String(), "/") + conf.PluginURLPath

	conf.MaxWebhookSize = 75 * 1024 * 1024 // 75Mb
	if newMattermostConfig.FileSettings.MaxFileSize != nil {
		conf.MaxWebhookSize = int(*newMattermostConfig.FileSettings.MaxFileSize)
	}

	conf.AWSAccessKey = os.Getenv(upaws.AccessEnvVar)
	conf.AWSSecretKey = os.Getenv(upaws.SecretEnvVar)
	conf.AWSRegion = upaws.Region()
	conf.AWSS3Bucket = upaws.S3BucketName()

	license := s.getMattermostLicense(log)
	conf.MattermostCloudMode = license != nil &&
		license.Features != nil &&
		license.Features.Cloud != nil &&
		*license.Features.Cloud
	if conf.MattermostCloudMode {
		log.Debugf("Detected Mattermost Cloud mode based on the license")
	}

	// On community.mattermost.com license is not suitable for checking, resort
	// to the presence of legacy environment variable to trigger it.
	legacyAccessKey := os.Getenv(upaws.DeprecatedCloudAccessEnvVar)
	if legacyAccessKey != "" {
		conf.MattermostCloudMode = true
		log.Debugf("Detected Mattermost Cloud mode based on the %s variable", upaws.DeprecatedCloudAccessEnvVar)
		conf.AWSAccessKey = legacyAccessKey
	}

	if conf.MattermostCloudMode {
		legacySecretKey := os.Getenv(upaws.DeprecatedCloudSecretEnvVar)
		if legacySecretKey != "" {
			conf.AWSSecretKey = legacySecretKey
		}
		if conf.AWSAccessKey == "" || conf.AWSSecretKey == "" {
			return nil, errors.New("access credentials for AWS must be set in Mattermost Cloud mode")
		}
	}

	if conf.AllowHTTPAppsOverride != nil {
		conf.AllowHTTPApps = *conf.AllowHTTPAppsOverride
	} else {
		conf.AllowHTTPApps = !conf.MattermostCloudMode || conf.DeveloperMode
	}

	return &conf, nil
}

func (s *service) Get() Config {
	s.lock.RLock()
	conf := s.conf
	s.lock.RUnlock()

	if conf == nil {
		return s.newConfig()
	}
	return *conf
}

func (s *service) MattermostAPI() *pluginapi.Client {
	return s.mm
}

func (s *service) I18N() *i18n.Bundle {
	return s.i18n
}

func (s *service) Telemetry() *telemetry.Telemetry {
	return s.telemetry
}

func (s *service) MattermostConfig() configservice.ConfigService {
	s.lock.RLock()
	mmconf := s.mattermostConfig
	s.lock.RUnlock()

	if mmconf == nil {
		mmconf = s.reloadMattermostConfig()
	}
	return &mattermostConfigService{
		mmconf: mmconf,
	}
}

func (s *service) reloadMattermostConfig() *model.Config {
	mmconf := s.mm.Configuration.GetConfig()

	s.lock.Lock()
	s.mattermostConfig = mmconf
	s.lock.Unlock()

	return mmconf
}

func (s *service) getMattermostLicense(log utils.Logger) *model.License {
	// GetLicense silently drops an RPC error
	// (https://github.com/mattermost/mattermost-server/blob/fc75b72bbabf7fabfad24b9e1e4c321ca9b9b7f1/plugin/client_rpc_generated.go#L864).
	// When running in Mattermost cloud we must not fall back to the on-prem mode, so in case we get a nil retry once.
	license := s.mm.System.GetLicense()
	if license == nil {
		license = s.mm.System.GetLicense()
		if license == nil {
			log.Debugf("failed to fetch license twice. May incorrectly default to on-prem mode")
		}
	}

	return license
}

func (s *service) Reconfigure(newStoredConfig StoredConfig, verbose bool, services ...Configurable) error {
	var log utils.Logger
	if verbose {
		// Use the old settings to log the processing of the new config.
		log = s.NewBaseLogger()
	} else {
		log = utils.NilLogger{}
	}

	clone, err := s.newInitializedConfig(newStoredConfig, log)
	if err != nil {
		return err
	}

	s.lock.Lock()
	s.conf = clone
	s.lock.Unlock()

	for _, service := range services {
		if err := service.Configure(*clone, log); err != nil {
			return errors.Wrapf(err, "failed to configure service %T", service)
		}
	}
	return nil
}

func (s *service) StoreConfig(sc StoredConfig, log utils.Logger) error {
	log.Debugf("Storing configuration: %d installed, %d listed apps, developer mode %t, allow http apps %t",
		len(sc.InstalledApps),
		len(sc.LocalManifests),
		sc.DeveloperModeOverride != nil && *sc.DeveloperModeOverride,
		sc.AllowHTTPAppsOverride != nil && *sc.AllowHTTPAppsOverride,
	)

	// Refresh computed values immediately, do not wait for OnConfigurationChanged
	err := s.Reconfigure(sc, false)
	if err != nil {
		return err
	}

	out := map[string]interface{}{}
	utils.Remarshal(&out, sc)

	// TODO test that SaveConfig will always cause OnConfigurationChange->c.Refresh
	return s.mm.Configuration.SavePluginConfig(out)
}

func (s *service) NewBaseLogger() utils.Logger {
	if s.Get().DeveloperMode {
		return utils.NewPluginLogger(s.mm, s)
	}
	return utils.NewPluginLogger(s.mm, nil)
}

func (s *service) GetLogConfig() utils.LogConfig {
	conf := s.Get()

	return utils.LogConfig{
		ChannelID:   conf.LogChannelID,
		Level:       zapcore.Level(conf.LogChannelLevel),
		BotUserID:   s.botUserID,
		IncludeJSON: conf.LogChannelJSON,
	}
}

func (s *service) SystemDefaultFlags() (bool, bool) {
	s.lock.RLock()
	mmconf := s.mattermostConfig
	s.lock.RUnlock()

	devMode := pluginapi.IsConfiguredForDevelopment(mmconf)
	allowHTTP := devMode || !s.Get().MattermostCloudMode
	return devMode, allowHTTP
}
