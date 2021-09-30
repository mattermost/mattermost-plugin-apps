package main

import (
	"testing"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

func TestOnActivate(t *testing.T) {
	testAPI := &plugintest.API{}
	p := NewPlugin(
		config.BuildConfig{
			Manifest:       manifest,
			BuildHash:      BuildHash,
			BuildHashShort: BuildHashShort,
			BuildDate:      BuildDate,
		},
	)

	p.API = testAPI

	testAPI.On("GetServerVersion").Return("5.30.1")
	testAPI.On("KVGet", "mmi_botid").Return([]byte("the_bot_id"), nil)

	username := "appsbot"
	displayName := "Mattermost Apps"
	description := "Mattermost Apps Registry and API proxy."
	testAPI.On("PatchBot", "the_bot_id", &model.BotPatch{
		Username:    &username,
		DisplayName: &displayName,
		Description: &description,
	}).Return(nil, nil)

	testAPI.On("GetBundlePath").Return("../", nil)

	testAPI.On("SetProfileImage", "the_bot_id", mock.AnythingOfType("[]uint8")).Return(nil)

	testAPI.On("LoadPluginConfiguration", mock.AnythingOfType("*config.StoredConfig")).Return(nil)

	listenAddress := "localhost:8065"
	siteURL := "http://" + listenAddress
	testAPI.On("GetConfig").Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL:       &siteURL,
			ListenAddress: &listenAddress,
		},
	})

	testAPI.On("GetLicense").Return(&model.License{
		Features:     &model.Features{},
		SkuShortName: "professional",
	})

	expectLog(testAPI, "LogDebug", 9)
	expectLog(testAPI, "LogInfo", 5)

	testAPI.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

	testAPI.On("PublishWebSocketEvent", "plugin_enabled", map[string]interface{}{"version": manifest.Version}, &model.WebsocketBroadcast{})

	err := p.OnActivate()
	require.NoError(t, err)
}

func TestOnDeactivate(t *testing.T) {
	testAPI := &plugintest.API{}
	p := NewPlugin(
		config.BuildConfig{
			Manifest:       manifest,
			BuildHash:      BuildHash,
			BuildHashShort: BuildHashShort,
			BuildDate:      BuildDate,
		},
	)

	p.API = testAPI

	mm := pluginapi.NewClient(p.API, p.Driver)
	p.conf = config.NewService(mm, p.BuildConfig, "the_bot_id", nil)

	testAPI.On("PublishWebSocketEvent", "plugin_disabled", map[string]interface{}{"version": manifest.Version}, &model.WebsocketBroadcast{})

	err := p.OnDeactivate()
	require.NoError(t, err)
}

func expectLog(testAPI *plugintest.API, logType string, numArgs int) {
	args := []interface{}{}
	for i := 0; i < numArgs; i++ {
		args = append(args, mock.Anything)
	}

	testAPI.On(logType, args...)
}
