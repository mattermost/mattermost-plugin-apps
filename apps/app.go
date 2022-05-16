// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"

	"github.com/mattermost/mattermost-server/v6/model"
)

// App describes an App installed on a Mattermost instance. App should be
// abbreviated as `app`.
type App struct {
	// Manifest contains the manifest data that the App was installed with. It
	// may differ from what is currently in the manifest store for the app's ID.
	Manifest

	// DeployType is the type of upstream that can be used to access the App.
	DeployType DeployType `json:"deploy_type,omitempty"`

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
	BotUserID   string `json:"bot_user_id,omitempty"`
	BotUsername string `json:"bot_username,omitempty"`

	// MattermostOAuth2 contains App's Mattermost OAuth2 credentials. An
	// Mattermost server OAuth2 app is created (or updated) when a Mattermost
	// App is installed on the instance.
	MattermostOAuth2 *model.OAuthApp `json:"mattermost_oauth2,omitempty"`

	// RemoteOAuth2 contains App's remote OAuth2 credentials. Use
	// appclient.StoreOAuth2App to update.
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

func DecodeCompatibleApp(data []byte) (app *App, err error) {
	defer func() {
		if app != nil {
			err = app.Validate()
			if err != nil {
				app = nil
			}
		}
	}()

	err = json.Unmarshal(data, &app)
	// If failed to decode as current version, opportunistically try as a
	// v0.7.x. There was no schema version before, this condition may need to be
	// updated in the future.
	if err != nil || app.SchemaVersion == "" {
		app7 := AppV0_7{}
		_ = json.Unmarshal(data, &app7)
		if from7 := app7.App(); from7 != nil {
			return from7, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return app, nil
}

// OAuth2App contains the setored settings for an "OAuth2 app" used by the App.
// It is used to describe the OAuth2 connections both to Mattermost, and
// optionally to a 3rd party remote system.
type OAuth2App struct {
	RemoteRootURL string `json:"remote_root_url,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	ClientSecret  string `json:"client_secret,omitempty"`

	// Data allows apps to store custom data in their OAuth configuration. A
	// frequent use case is storing app-level service account credentials.
	Data interface{} `json:"data,omitempty"`
}

// ListedApp is a Mattermost App listed in the Marketplace containing metadata.
type ListedApp struct {
	Manifest  Manifest                 `json:"manifest"`
	Installed bool                     `json:"installed"`
	Enabled   bool                     `json:"enabled"`
	IconURL   string                   `json:"icon_url,omitempty"`
	Labels    []model.MarketplaceLabel `json:"labels,omitempty"`
}
