package main

import (
	"path/filepath"
	"testing"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-api/i18n"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v6/plugin/plugintest/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func TestOnActivate(t *testing.T) {
	testAPI := &plugintest.API{}
	p := NewPlugin(manifest)
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

	testAPI.On("KVSetWithOptions", "mutex_mmi_bot_ensure", []byte{0x1}, model.PluginKVSetOptions{Atomic: true, OldValue: nil, ExpireInSeconds: 15}).Return(true, nil)
	testAPI.On("KVSetWithOptions", "mutex_mmi_bot_ensure", []byte(nil), model.PluginKVSetOptions{Atomic: false, OldValue: nil, ExpireInSeconds: 0}).Return(true, nil)

	testAPI.On("GetBundlePath").Return("../", nil)

	testAPI.On("SetProfileImage", "the_bot_id", mock.AnythingOfType("[]uint8")).Return(nil)

	testAPI.On("GetPluginConfig").Return(map[string]any{})

	listenAddress := "localhost:8065"
	siteURL := "http://" + listenAddress + "/subpath"
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

	expectLog(testAPI, "LogDebug", 13)
	expectLog(testAPI, "LogInfo", 5)
	expectLog(testAPI, "LogError", 3)

	testAPI.On("RegisterCommand", mock.AnythingOfType("*model.Command")).Return(nil)

	testAPI.On("PublishWebSocketEvent", "plugin_enabled", map[string]interface{}{"version": manifest.Version}, &model.WebsocketBroadcast{})

	err := p.OnActivate()
	require.NoError(t, err)
}

func TestOnDeactivate(t *testing.T) {
	testAPI := &plugintest.API{}
	p := NewPlugin(manifest)

	listenAddress := "localhost:8065"
	siteURL := "http://" + listenAddress + "/subpath"
	p.API = testAPI
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
	testAPI.On("GetPluginConfig").Return(map[string]any{})

	testAPI.On("GetBundlePath").Return("/", nil)
	i18nBundle, _ := i18n.InitBundle(testAPI, filepath.Join("assets", "i18n"))

	mm := pluginapi.NewClient(p.API, p.Driver)
	confService, err := config.NewService(mm, manifest, "the_bot_id", nil, i18nBundle, utils.NewTestLogger())
	require.NoError(t, err)
	p.conf = confService

	testAPI.On("PublishWebSocketEvent", "plugin_disabled", map[string]interface{}{"version": manifest.Version}, &model.WebsocketBroadcast{})

	err = p.OnDeactivate()
	require.NoError(t, err)
}

func expectLog(testAPI *plugintest.API, logType string, numArgs int) {
	args := []interface{}{}
	for i := 0; i < numArgs; i++ {
		args = append(args, mock.Anything)
	}

	testAPI.On(logType, args...)
}
