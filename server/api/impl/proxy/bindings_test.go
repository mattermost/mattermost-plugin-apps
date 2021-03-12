// +build !e2e

package proxy

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/api/mock_api"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
)
const (
	testAppID1  = "test-appID1"
	testUserID  = "test-userID"
	testChannelID = "test-channelID"
)

func buildTestAppBindings() map[apps.AppID][]*apps.Binding {
	bindingsForApp := map[apps.AppID][]*apps.Binding{}

	n := "simple"
	appID := apps.AppID(testAppID1)
	bindingsApp1 := []*apps.Binding{
		{
			AppID:    appID,
			Location: "/channel_header",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: false,
			DependsOnUser: false,
		},
		{
			AppID:    appID,
			Location: "/command",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: true,
			DependsOnUser: false,
		},
		{
			AppID:    appID,
			Location: "/post_menu",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: false,
			DependsOnUser: true,
		},
		{
			AppID:    appID,
			Location: "/in_post",
			Bindings: []*apps.Binding{
				{
					AppID:    appID,
					Location: apps.Location(fmt.Sprintf("id-%s", n)),
					Hint:     fmt.Sprintf("hint-%s", n),
				},
			},
			DependsOnChannel: true,
			DependsOnUser: true,
		},
	}

	bindingsForApp[appID] = bindingsApp1

	return bindingsForApp
}

func newTestProxyForBindings(testAppBindings map[apps.AppID][]*apps.Binding) (*Proxy, *plugintest.API) {
	testAPI := &plugintest.API{}
	mm := pluginapi.NewClient(testAPI)

	p := &Proxy{
		mm:      mm,
	}

	return p, testAPI
}

func TestGetAllBindings(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	key1 := p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)
	key2 := p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)
	key3 := p.CacheBuildKey(KEY_ALL_USERS, testChannelID)
	key4 := p.CacheBuildKey(testUserID, testChannelID)

	bindingBytes := []byte{}
	bindingsBytes := [][]byte{}
	bindingBytes, _ = json.Marshal(&appBindings[appID][0])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key1).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&appBindings[appID][1])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key2).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&appBindings[appID][2])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key3).Return(bindingsBytes, nil)

	bindingsBytes = [][]byte{}
	bindingBytes, _ = json.Marshal(&appBindings[appID][3])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key4).Return(bindingsBytes, nil)

	outBindings, err := p.CacheGetAll(context1, appID)
	require.Equal(t, outBindings, appBindings[appID])

	assert.NoError(t, err)
}

func TestGetBindings(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	key2 := p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)

	bindingsBytes := [][]byte{}
	bindingBytes, _ := json.Marshal(&appBindings[appID][1])
	bindingsBytes = append(bindingsBytes, bindingBytes)
	api.On("AppsCacheGet", string(appID), key2).Return(bindingsBytes, nil)

	outBindings, err := p.CacheGet(context1, appID, key2)
	bindings := make([]*apps.Binding,0)
	bindings = append(bindings, appBindings[appID][1])
	require.Equal(t, outBindings, bindings)

	assert.NoError(t, err)
}

func TestDeleteBindingsForApp(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	key1 := p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)

	api.On("AppsCacheDelete", string(appID), key1).Return(nil)
	err := p.CacheDelete(appID, key1)

	assert.NoError(t, err)
}

func TestDeleteAllBindings(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	api.On("AppsCacheDeleteAll", string(appID)).Return(nil)
	err := p.CacheDeleteAll(appID)

	assert.NoError(t, err)
}

func TestDeleteAllBindingsForAllApps(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)

	s := mock_api.NewMockStore(controller)
	appStore := mock_api.NewMockAppStore(controller)
	s.EXPECT().App().Return(appStore)
	p.store = s

	appList := []*apps.App{}
	for k, _ := range appBindings {
		tapp := &apps.App{
			Manifest: &apps.Manifest{
				AppID:              apps.AppID(k),
				Type:               apps.AppTypeBuiltin,
			},
		}
		appList = append(appList, tapp)
	}
	appStore.EXPECT().GetAll().Return(appList)

	api.On("AppsCacheDeleteAll", string(appID)).Return(nil)
	errors := p.CacheDeleteAllApps()

	assert.Equal(t, len(errors), 0)
}

func TestSetBindings(t *testing.T) {
	appBindings := buildTestAppBindings()
	controller := gomock.NewController(t)
	defer controller.Finish()
	p, api := newTestProxyForBindings(appBindings)
	defer api.AssertExpectations(t)

	appID := apps.AppID(testAppID1)
	context1 := &apps.Context {
		AppID : appID,
		ActingUserID: testUserID,
		UserID: testUserID,
		ChannelID: testChannelID,
	}

	var key string
	bindingsMap := make(map[string][][]byte, len(appBindings[appID]))

	for _, binding := range appBindings[appID] {
		if !binding.DependsOnUser && !binding.DependsOnChannel {
			key = p.CacheBuildKey(KEY_ALL_USERS, KEY_ALL_CHANNELS)
		} else if binding.DependsOnUser && !binding.DependsOnChannel {
			key = p.CacheBuildKey(testUserID, KEY_ALL_CHANNELS)
		} else if !binding.DependsOnUser && binding.DependsOnChannel {
			key = p.CacheBuildKey(KEY_ALL_USERS, testChannelID)
		} else {
			key = p.CacheBuildKey(testUserID, testChannelID)
		}
		valueBytes, _ := json.Marshal(&binding)

		bindingsMap[key] = make([][]byte, 0)
		bindingsForKey := bindingsMap[key]
		bindingsForKey = append(bindingsForKey, valueBytes)
		bindingsMap[key] = bindingsForKey

		api.On("AppsCacheSet", string(appID), bindingsMap).Return(nil)
	}

	err := p.CacheSet(context1, appID, appBindings[appID])

	assert.NoError(t, err)
}