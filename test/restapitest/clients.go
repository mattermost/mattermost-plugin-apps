// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
)

type TestClientPP struct {
	*appclient.ClientPP
	UserID string
}

func (th *Helper) InitClients() {
	init := func(loginf func(*model.Client4)) *appclient.ClientPP {
		c := th.ServerTestHelper.CreateClient()
		loginf(c)

		cpp := th.CreateUnauthenticatedClientPP()
		cpp.AuthToken = c.AuthToken
		cpp.AuthType = c.AuthType
		return cpp
	}

	th.UserClientPP = init(th.ServerTestHelper.LoginBasicWithClient)
	th.User2ClientPP = init(th.ServerTestHelper.LoginBasic2WithClient)
	th.SystemAdminClientPP = init(th.ServerTestHelper.LoginSystemAdminWithClient)

	th.LocalClientPP = th.CreateLocalClient(*th.ServerTestHelper.App.Config().ServiceSettings.LocalModeSocketLocation)
}

func (th *Helper) CreateUnauthenticatedClientPP() *appclient.ClientPP {
	cfg := th.ServerTestHelper.App.Config()
	siteURL, err := url.Parse(*cfg.ServiceSettings.SiteURL)
	require.NoError(th, err)
	url := fmt.Sprintf("http://localhost:%v", th.ServerTestHelper.App.Srv().ListenAddr.Port) + siteURL.Path
	return appclient.NewAppsPluginAPIClient(url)
}

func (th *Helper) CreateLocalClient(socketPath string) *appclient.ClientPP {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	client := appclient.NewAppsPluginAPIClient("http://_" + model.APIURLSuffix)
	client.HTTPClient = httpClient

	return client
}

func (th *Helper) TestForUser(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"UserClientPP", func(t *testing.T) {
		f(t, th.UserClientPP)
	})
}

func (th *Helper) TestForSystemAdmin(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"SystemAdminClientPP", func(t *testing.T) {
		f(t, th.SystemAdminClientPP)
	})
}

func (th *Helper) TestForLocal(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"LocalClientPP", func(t *testing.T) {
		f(t, th.LocalClientPP)
	})
}

func (th *Helper) TestForUserAndSystemAdmin(t *testing.T, f func(*testing.T, *appclient.ClientPP), name ...string) {
	th.TestForUser(t, f)
	th.TestForSystemAdmin(t, f)
}
