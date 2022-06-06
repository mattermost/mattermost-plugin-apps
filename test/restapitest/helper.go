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

type Caller func(apps.AppID, apps.CallRequest) *apps.CallResponse

func NewHelper(t *testing.T, apps ...*goapp.App) *Helper {
	require := require.New(t)
	// Check environment
	require.NotEmpty(os.Getenv("MM_SERVER_PATH"),
		"MM_SERVER_PATH is not set, please set it to the path of your mattermost-server clone")

	// Unset SiteURL, just in case it's set
	err := os.Unsetenv("MM_SERVICESETTINGS_SITEURL")
	require.NoError(err)

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
	require := require.New(th)

	th.verifyExpandedContext(level, app, asSystemAdmin, expected.ExpandedContext, got.ExpandedContext)

	expected.ExpandedContext = apps.ExpandedContext{}
	got.ExpandedContext = apps.ExpandedContext{}
	require.EqualValues(expected, got)
}

func (th *Helper) verifyExpandedContext(level apps.ExpandLevel, app *apps.App, asSystemAdmin bool, expected, got apps.ExpandedContext) {
	siteURL := *th.ServerTestHelper.Server.Config().ServiceSettings.SiteURL
	appPath := "/plugins/com.mattermost.apps/apps/" + string(app.AppID)
	require := require.New(th)
	require.Equal(siteURL, got.MattermostSiteURL)
	require.Equal(appPath, got.AppPath)
	require.Equal(app.BotUserID, got.BotUserID)
	require.Equal(expected.DeveloperMode, got.DeveloperMode)

	require.NotEmpty(got.BotAccessToken)
	expected.BotAccessToken, got.BotAccessToken = "", ""
	if level != apps.ExpandAll {
		require.Empty(got.ActingUserAccessToken)
	} else {
		require.NotEmpty(got.ActingUserAccessToken)
	}
	expected.ActingUserAccessToken, got.ActingUserAccessToken = "", ""

	if level == apps.ExpandNone {
		// make sure nothing else is set.
		require.EqualValues(apps.ExpandedContext{
			MattermostSiteURL: siteURL,
			AppPath:           appPath,
			BotUserID:         app.BotUserID,
			DeveloperMode:     expected.DeveloperMode,
		}, got)
		return
	}

	require.Equal(expected.Locale, got.Locale)
	require.EqualValues(expected.OAuth2, got.OAuth2)
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
	require := require.New(th)
	if expected == nil || expected.Id == "" {
		require.Empty(th, got)
		return
	}
	require.NotNil(th, got)

	switch level {
	case apps.ExpandNone:
		require.Empty(got)
		return

	case apps.ExpandID:
		// Make sure nothing else is set.
		expected = &model.User{Id: expected.Id}

	case apps.ExpandSummary:
		expected = &model.User{
			Id:             expected.Id,
			Email:          expected.Email,
			Position:       expected.Position,
			FirstName:      expected.FirstName,
			LastName:       expected.LastName,
			Locale:         expected.Locale,
			Nickname:       expected.Nickname,
			Roles:          expected.Roles,
			Timezone:       expected.Timezone,
			Username:       expected.Username,
			IsBot:          expected.IsBot,
			BotDescription: expected.BotDescription,
			DeleteAt:       expected.DeleteAt,
		}

	case apps.ExpandAll:
		expected = &model.User{
			Id:             expected.Id,
			Email:          expected.Email,
			Position:       expected.Position,
			FirstName:      expected.FirstName,
			LastName:       expected.LastName,
			Locale:         expected.Locale,
			Nickname:       expected.Nickname,
			Roles:          expected.Roles,
			Timezone:       expected.Timezone,
			Username:       expected.Username,
			IsBot:          expected.IsBot,
			BotDescription: expected.BotDescription,
			DeleteAt:       expected.DeleteAt,
			Props:          expected.Props,
			CreateAt:       expected.CreateAt,
		}

		// Zero out "got" fields that are ignored for the purpose of verification.
		comp := *got
		got = &comp
		got.UpdateAt = 0
		got.AuthData = nil
		got.AuthService = ""
		got.EmailVerified = false
		got.NotifyProps = nil
		got.LastPasswordUpdate = 0
		got.LastPictureUpdate = 0
		got.FailedAttempts = 0
		got.MfaActive = false
		got.MfaSecret = ""
		got.RemoteId = nil
		got.LastActivityAt = 0
		got.BotLastIconUpdate = 0
		got.TermsOfServiceId = ""
		got.TermsOfServiceCreateAt = 0
		got.DisableWelcomeEmail = false
		got.Password = ""

		// empty props comes differently from different places; normalize for
		// comparison.
		if expected.Props == nil {
			expected.Props = model.StringMap{}
		}
		if got.Props == nil {
			got.Props = model.StringMap{}
		}
	}

	sanitizeMap := map[string]bool{
		"email":    true,
		"fullname": true,
	}
	expected.Sanitize(sanitizeMap)
	got.Sanitize(sanitizeMap)
	require.EqualValues(expected, got)
}

func (th *Helper) requireEqualApp(level apps.ExpandLevel, asSystemAdmin bool, expected, got *apps.App) {
	switch level {
	case apps.ExpandSummary:
		app := &apps.App{
			Manifest: apps.Manifest{
				AppID:   expected.AppID,
				Version: expected.Version,
			},
			BotUserID:   expected.BotUserID,
			BotUsername: expected.BotUsername,
		}
		require.EqualValues(th, app, got)

	case apps.ExpandAll:
		app := &apps.App{
			Manifest: apps.Manifest{
				AppID:   expected.AppID,
				Version: expected.Version,
			},
			BotUserID:     expected.BotUserID,
			BotUsername:   expected.BotUsername,
			DeployType:    expected.DeployType,
			WebhookSecret: expected.WebhookSecret,
		}
		// Only sysadmins get the webhook secret expanded.
		if !asSystemAdmin {
			app.WebhookSecret = ""
		}
		require.EqualValues(th, app, got)

	default:
		require.Nil(th, got)
	}
}

func (th *Helper) requireEqualChannel(level apps.ExpandLevel, expected, got *model.Channel) {
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualChannelMember(level apps.ExpandLevel, expected, got *model.ChannelMember) {
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualTeam(level apps.ExpandLevel, expected, got *model.Team) {
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualTeamMember(level apps.ExpandLevel, expected, got *model.TeamMember) {
	require.EqualValues(th, expected, got)
}

func (th *Helper) requireEqualPost(level apps.ExpandLevel, expected, got *model.Post) {
	require.EqualValues(th, expected, got)
}
