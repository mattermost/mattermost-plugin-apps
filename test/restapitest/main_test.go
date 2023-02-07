// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/testlib"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	var options = testlib.HelperOptions{
		EnableStore:     true,
		EnableResources: false,
	}

	mainHelper = testlib.NewMainHelperWithOptions(&options)
	defer mainHelper.Close()

	api4.SetMainHelper(mainHelper)

	mainHelper.Main(m)
}

func TestRESTAPI(t *testing.T) {
	th := NewHelper(t)

	for name, testF := range map[string]func(*Helper){
		"calls":         testCalls,
		"bindings":      testBindings,
		"echo":          testEcho,
		"KV":            testKV,
		"OAuth2":        testOAuth2,
		"webhook_auth":  testWebhookAuth,
		"webhook_path":  testWebhookPath,
		"subscriptions": testSubscriptions,
		"static":        testStatic,
		"notify":        testNotify,
		"uninstall":     testUninstall,
	} {
		th.CleanRun(name, testF)
	}
}
