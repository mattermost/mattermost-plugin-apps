// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package restapitest

import (
	"fmt"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/v6/api4"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

// Note: run
// set export MM_SERVER_PATH="<go path>/src/github.com/mattermost/mattermost-server"
// command (or equivalent) before running the tests
var pluginID = "com.mattermost.apps"

type Helper struct {
	*testing.T
	ServerTestHelper *api4.TestHelper

	UserClientPP        *appclient.ClientPP
	User2ClientPP       *appclient.ClientPP
	SystemAdminClientPP *appclient.ClientPP
	LocalClientPP       *appclient.ClientPP
}

type TestFunc func(*Helper)
type Caller func(apps.AppID, apps.CallRequest) *apps.CallResponse

func NewHelper(t *testing.T, apps ...*goapp.App) *Helper {
	// Check environment
	require.NotEmpty(t, os.Getenv("MM_SERVER_PATH"),
		"MM_SERVER_PATH is not set, please set it to the path of your mattermost-server clone")

	// Unset SiteURL, just in case it's set
	err := os.Unsetenv("MM_SERVICESETTINGS_SITEURL")
	require.NoError(t, err)

	// Setup Mattermost server (helper)
	serverTestHelper := api4.Setup(t)
	serverTestHelper.InitBasic()
	port := serverTestHelper.Server.ListenAddr.Port
	serverTestHelper.App.UpdateConfig(func(cfg *model.Config) {
		// Need to create plugin and app bots.
		*cfg.ServiceSettings.EnableBotAccountCreation = true

		// Need to create and use OAuth2 apps.
		*cfg.ServiceSettings.EnableOAuthServiceProvider = true

		// Need to make requests to other local servers (apps).
		*cfg.ServiceSettings.AllowedUntrustedInternalConnections = "127.0.0.1"

		// Update the server own address, as we know it.
		*cfg.ServiceSettings.SiteURL = fmt.Sprintf("http://localhost:%d", port)
		*cfg.ServiceSettings.ListenAddress = fmt.Sprintf(":%d", port)

		// Enable developer mode for better logging
		*cfg.ServiceSettings.EnableDeveloper = true
		*cfg.ServiceSettings.EnableTesting = true
	})

	// Create the helper and register for cleanup.
	th := &Helper{
		T:                t,
		ServerTestHelper: serverTestHelper,
	}
	t.Cleanup(th.TearDown)

	th.InitClients()
	th.InstallAppsPlugin()
	for _, a := range apps {
		th.InstallAppWithCleanup(a)
	}
	return th
}

func (th *Helper) TearDown() {
	th.ServerTestHelper.TearDown()
}

func (th *Helper) Run(name string, f func(th *Helper)) bool {
	return th.T.Run(name, func(t *testing.T) {
		h := *th
		h.T = t
		f(&h)
	})
}

func respond(text string, err error) apps.CallResponse {
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	return apps.NewTextResponse(text)
}

func (th *Helper) verifyContext(level apps.ExpandLevel, app *apps.App, asSystemAdmin bool, expected, got apps.Context) {
	th.verifyExpandedContext(level, app, asSystemAdmin, expected.ExpandedContext, got.ExpandedContext)

	expected.ExpandedContext = apps.ExpandedContext{}
	got.ExpandedContext = apps.ExpandedContext{}
	require.EqualValues(th, expected, got)
}

func (th *Helper) verifyExpandedContext(level apps.ExpandLevel, app *apps.App, asSystemAdmin bool, expected, got apps.ExpandedContext) {
	siteURL := *th.ServerTestHelper.Server.Config().ServiceSettings.SiteURL
	appPath := "/plugins/com.mattermost.apps/apps/" + string(app.AppID)
	require.Equal(th, siteURL, got.MattermostSiteURL)
	require.Equal(th, appPath, got.AppPath)
	require.Equal(th, app.BotUserID, got.BotUserID)

	// The dev mode is always set in the test.
	require.Equal(th, true, got.DeveloperMode)

	require.NotEmpty(th, got.BotAccessToken)
	expected.BotAccessToken, got.BotAccessToken = "", ""
	if level != apps.ExpandAll {
		require.Empty(th, got.ActingUserAccessToken)
	} else {
		require.NotEmpty(th, got.ActingUserAccessToken)
	}
	expected.ActingUserAccessToken, got.ActingUserAccessToken = "", ""

	if level == apps.ExpandNone {
		// make sure nothing else is set.
		require.EqualValues(th, apps.ExpandedContext{
			MattermostSiteURL: siteURL,
			AppPath:           appPath,
			BotUserID:         app.BotUserID,
			DeveloperMode:     true,
		}, got)
		return
	}

	require.Equal(th, expected.Locale, got.Locale)
	require.EqualValues(th, expected.OAuth2, got.OAuth2)
	th.requireEqualApp(level, asSystemAdmin, expected.App, got.App)
	th.requireEqualUser(level, expected.ActingUser, got.ActingUser)
	th.requireEqualChannel(level, expected.Channel, got.Channel)
	th.requireEqualChannelMember(level, expected.ChannelMember, got.ChannelMember)
	th.requireEqualTeam(level, expected.Team, got.Team)
	th.requireEqualTeamMember(level, expected.TeamMember, got.TeamMember)
	th.requireEqualPost(level, expected.Post, got.Post)
	th.requireEqualPost(level, expected.RootPost, got.RootPost)
	th.requireEqualUser(level, expected.User, got.User)
}

func (th *Helper) requireEqualUser(level apps.ExpandLevel, expected, got *model.User) {
	if expected == nil || expected.Id == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	// Zero out fields that are ignored for the purpose of verification.
	comparable := func(user *model.User) *model.User {
		if user == nil {
			return nil
		}
		clone := *user
		clone.UpdateAt = 0
		clone.AuthData = nil
		clone.AuthService = ""
		clone.EmailVerified = false
		clone.NotifyProps = nil
		clone.LastPasswordUpdate = 0
		clone.LastPictureUpdate = 0
		clone.FailedAttempts = 0
		clone.MfaActive = false
		clone.MfaSecret = ""
		clone.RemoteId = nil
		clone.LastActivityAt = 0
		clone.BotLastIconUpdate = 0
		clone.TermsOfServiceId = ""
		clone.TermsOfServiceCreateAt = 0
		clone.DisableWelcomeEmail = false
		clone.Password = ""
		if clone.Props == nil {
			clone.Props = model.StringMap{}
		}

		clone.Sanitize(map[string]bool{
			"email":    true,
			"fullname": true,
		})
		return &clone
	}

	expected = apps.StripUser(expected, level)
	require.EqualValues(th, comparable(expected), comparable(got))
}

func (th *Helper) requireEqualApp(level apps.ExpandLevel, asSystemAdmin bool, expected, got *apps.App) {
	app := expected.Strip(level)
	// Only sysadmins get the webhook secret expanded.
	if !asSystemAdmin && app != nil {
		app.WebhookSecret = ""
	}
	require.EqualValues(th, app, got)
}

func (th *Helper) requireEqualChannel(level apps.ExpandLevel, expected, got *model.Channel) {
	if expected == nil || expected.Id == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	// Zero out fields that are ignored for the purpose of verification.
	comparable := func(channel *model.Channel) *model.Channel {
		if channel == nil {
			return nil
		}
		clone := *channel
		clone.UpdateAt = 0
		if clone.Props == nil {
			clone.Props = map[string]interface{}{}
		}
		return &clone
	}

	expected = apps.StripChannel(expected, level)
	require.EqualValues(th, comparable(expected), comparable(got))
}

func (th *Helper) requireEqualChannelMember(level apps.ExpandLevel, expected, got *model.ChannelMember) {
	if expected == nil || expected.UserId == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	expected = apps.StripChannelMember(expected, level)
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualTeam(level apps.ExpandLevel, expected, got *model.Team) {
	if expected == nil || expected.Id == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	expected = apps.StripTeam(expected, level)
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualTeamMember(level apps.ExpandLevel, expected, got *model.TeamMember) {
	if expected == nil || expected.UserId == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	expected = apps.StripTeamMember(expected, level)
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualPost(level apps.ExpandLevel, expected, got *model.Post) {
	if expected == nil || expected.Id == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	if level == apps.ExpandNone {
		require.Empty(th, got)
		return
	}

	expected = apps.StripPost(expected, level)
	require.EqualValues(th, expected, got)
}
