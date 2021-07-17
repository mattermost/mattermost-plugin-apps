package apps

import (
	"unicode"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// AppID is a globally unique identifier that represents a Mattermost App.
// An AppID is restricted to no more than 32 ASCII letters, numbers, '-', or '_'.
type AppID string

// AppVersion is the version of a Mattermost App. AppVersion is expected to look
// like "v00_00_000".
type AppVersion string

// AppType is the type of an app: http, aws_lambda, or builtin.
type AppType string

// App describes an App installed on a Mattermost instance. App should be
// abbreviated as `app`.
type App struct {
	// Manifest contains the manifest data that the App was installed with. It
	// may differ from what is currently in the manifest store for the app's ID.
	Manifest

	// Disabled is set to true if the app is disabled. Disabling an app does not
	// erase any of it's data.
	Disabled bool `json:"disabled,omitempty"`

	// Secret is used to issue JWT when sending requests to HTTP apps.
	Secret string `json:"secret,omitempty"`

	// WebhookSecret is used to validate an incoming webhook secret.
	WebhookSecret string `json:"webhook_secret,omitempty"`

	// App's Mattermost Bot User credentials. An Mattermost server Bot Account
	// is created (or updated) when a Mattermost App is installed on the
	// instance.
	BotUserID        string `json:"bot_user_id,omitempty"`
	BotUsername      string `json:"bot_username,omitempty"`
	BotAccessToken   string `json:"bot_access_token,omitempty"`
	BotAccessTokenID string `json:"bot_access_token_id,omitempty"`

	// Trusted means that Mattermost will issue the Apps' users their tokens as
	// needed, without asking for the user's consent.
	Trusted bool `json:"trusted,omitempty"`

	// MattermostOAuth2 contains App's Mattermost OAuth2 credentials. An
	// Mattermost server OAuth2 app is created (or updated) when a Mattermost
	// App is installed on the instance.
	MattermostOAuth2 OAuth2App `json:"mattermost_oauth2,omitempty"`

	// RemoteOAuth2 contains App's remote OAuth2 credentials. Use
	// mmclient.StoreOAuth2App to update.
	RemoteOAuth2 OAuth2App `json:"remote_oauth2,omitempty"`

	// In V1, GrantedPermissions are simply copied from RequestedPermissions
	// upon the sysadmin's consent, during installing the App.
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`

	// GrantedLocations contains the list of top locations that the application
	// is allowed to bind to.
	//
	// In V1, GrantedLocations are simply copied from RequestedLocations upon
	// the sysadmin's consent, during installing the App.
	GrantedLocations Locations `json:"granted_locations,omitempty"`
}

// OAuth2App contains the setored settings for an "OAuth2 app" used by the App.
// It is used to describe the OAuth2 connections both to Mattermost, and
// optionally to a 3rd party remote system.
type OAuth2App struct {
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// ListedApp is a Mattermost App listed in the Marketplace containing metadata.
type ListedApp struct {
	Manifest  *Manifest                `json:"manifest"`
	Installed bool                     `json:"installed"`
	Enabled   bool                     `json:"enabled"`
	IconURL   string                   `json:"icon_url,omitempty"`
	Labels    []model.MarketplaceLabel `json:"labels,omitempty"`
}

type AppMetadataForClient struct {
	BotUserID   string `json:"bot_user_id,omitempty"`
	BotUsername string `json:"bot_username,omitempty"`
}

const (
	MinAppIDLength = 3
	MaxAppIDLength = 32
)

func (id AppID) IsValid() error {
	if len(id) < MinAppIDLength {
		return utils.NewInvalidError("appID %s too short, should be %d bytes", id, MinAppIDLength)
	}

	if len(id) > MaxAppIDLength {
		return utils.NewInvalidError("appID %s too long, should be %d bytes", id, MaxAppIDLength)
	}

	for _, c := range id {
		if unicode.IsLetter(c) {
			continue
		}

		if unicode.IsNumber(c) {
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			continue
		}

		return utils.NewInvalidError("invalid character '%c' in appID %q", c, id)
	}

	return nil
}

const VersionFormat = "v00_00_000"

func (v AppVersion) IsValid() error {
	if len(v) > len(VersionFormat) {
		return utils.NewInvalidError("version %s too long, should be in %s format", v, VersionFormat)
	}

	for _, c := range v {
		if unicode.IsLetter(c) {
			continue
		}

		if unicode.IsNumber(c) {
			continue
		}

		if c == '-' || c == '_' || c == '.' {
			continue
		}

		return utils.NewInvalidError("invalid character '%c' in appVersion", c)
	}

	return nil
}

const (
	// HTTP app (default). All communications are done via HTTP requests. Paths
	// for both functions and static assets are appended to RootURL "as is".
	// Mattermost authenticates to the App with an optional shared secret based
	// JWT.
	AppTypeHTTP AppType = "http"

	// AWS Lambda app. All functions are called via AWS Lambda "Invoke" API,
	// using path mapping provided in the app's manifest. Static assets are
	// served out of AWS S3, using the "Download" method. Mattermost
	// authenticates to AWS, no authentication to the App is necessary.
	AppTypeAWSLambda AppType = "aws_lambda"

	AppTypeKubeless AppType = "kubeless"

	// Builtin app. All functions and resources are served by directly invoking
	// go functions. No manifest, no Mattermost to App authentication are
	// needed.
	AppTypeBuiltin AppType = "builtin"
)

func (at AppType) IsValid() error {
	switch at {
	case AppTypeHTTP, AppTypeAWSLambda, AppTypeBuiltin, AppTypeKubeless:
		return nil
	default:
		return utils.NewInvalidError("%s is not a valid app type", at)
	}
}
