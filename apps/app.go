package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

// App describes an App installed on a Mattermost instance. App should be
// abbreviated as `app`.
type App struct {
	// Manifest contains the manifest data that the App was installed with. It
	// may differ from what is currently in the manifest store for the app's ID.
	Manifest

	// Type is the type of the App used in for this installation. An app may be available as a
	Type AppType `json:"type"`

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
