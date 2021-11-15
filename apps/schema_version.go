// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

type AppV0_7 struct {
	ManifestV0_7

	Disabled           bool        `json:"disabled,omitempty"`
	Secret             string      `json:"secret,omitempty"`
	WebhookSecret      string      `json:"webhook_secret,omitempty"`
	BotUserID          string      `json:"bot_user_id,omitempty"`
	BotUsername        string      `json:"bot_username,omitempty"`
	BotAccessToken     string      `json:"bot_access_token,omitempty"`
	BotAccessTokenID   string      `json:"bot_access_token_id,omitempty"`
	Trusted            bool        `json:"trusted,omitempty"`
	MattermostOAuth2   OAuth2App   `json:"mattermost_oauth2,omitempty"`
	RemoteOAuth2       OAuth2App   `json:"remote_oauth2,omitempty"`
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`
	GrantedLocations   Locations   `json:"granted_locations,omitempty"`
}

func (a7 AppV0_7) App() *App {
	m := a7.ManifestV0_7.Manifest()
	if m == nil {
		return nil
	}

	return &App{
		Manifest:   *m,
		DeployType: DeployType(m.v7AppType),

		Disabled:           a7.Disabled,
		Secret:             a7.Secret,
		WebhookSecret:      a7.WebhookSecret,
		BotUserID:          a7.BotUserID,
		BotUsername:        a7.BotUsername,
		RemoteOAuth2:       a7.RemoteOAuth2,
		GrantedPermissions: a7.GrantedPermissions,
		GrantedLocations:   a7.GrantedLocations,
	}
}

// ManifestV7 is the v0.7.0 version of Manifest, see
// https://github.com/mattermost/mattermost-plugin-apps/blob/7069695da0c3d14ff449dd0abf8d6cdb261c0df9/apps/manifest.go#L21.
type ManifestV0_7 struct {
	AppID                AppID               `json:"app_id"`
	AppType              string              `json:"app_type"`
	Version              AppVersion          `json:"version"`
	HomepageURL          string              `json:"homepage_url"`
	DisplayName          string              `json:"display_name,omitempty"`
	Description          string              `json:"description,omitempty"`
	Icon                 string              `json:"icon,omitempty"`
	Bindings             *Call               `json:"bindings,omitempty"`
	OnInstall            *Call               `json:"on_install,omitempty"`
	OnVersionChanged     *Call               `json:"on_version_changed,omitempty"`
	OnUninstall          *Call               `json:"on_uninstall,omitempty"`
	OnDisable            *Call               `json:"on_disable,omitempty"`
	OnEnable             *Call               `json:"on_enable,omitempty"`
	GetOAuth2ConnectURL  *Call               `json:"get_oauth2_connect_url,omitempty"`
	OnOAuth2Complete     *Call               `json:"on_oauth2_complete,omitempty"`
	RequestedPermissions Permissions         `json:"requested_permissions,omitempty"`
	RequestedLocations   Locations           `json:"requested_locations,omitempty"`
	HTTPRootURL          string              `json:"root_url,omitempty"`
	AWSLambda            []AWSLambdaFunction `json:"aws_lambda,omitempty"`
	PluginID             string              `json:"plugin_id,omitempty"`
}

func (m7 ManifestV0_7) Manifest() *Manifest {
	m := Manifest{}

	switch {
	case m7.AppType == string(DeployHTTP) && m7.HTTPRootURL != "":
		m.HTTP = &HTTP{
			RootURL: m7.HTTPRootURL,
		}
	case m7.AppType == string(DeployAWSLambda) && len(m7.AWSLambda) > 0:
		m.AWSLambda = &AWSLambda{
			Functions: m7.AWSLambda,
		}
	case m7.AppType == string(DeployPlugin) && m7.PluginID != "":
		m.Plugin = &Plugin{
			PluginID: m7.PluginID,
		}
	default:
		// invalid, ignore.
		return nil
	}

	m.AppID = m7.AppID
	m.Version = m7.Version
	m.HomepageURL = m7.HomepageURL
	m.DisplayName = m7.DisplayName
	m.Description = m7.Description
	m.Icon = m7.Icon
	m.Bindings = m7.Bindings
	m.OnInstall = m7.OnInstall
	m.OnVersionChanged = m7.OnVersionChanged
	m.OnUninstall = m7.OnUninstall
	m.OnDisable = m7.OnDisable
	m.OnEnable = m7.OnEnable
	m.GetOAuth2ConnectURL = m7.GetOAuth2ConnectURL
	m.OnOAuth2Complete = m7.OnOAuth2Complete
	m.RequestedPermissions = m7.RequestedPermissions
	m.RequestedLocations = m7.RequestedLocations

	m.v7AppType = m7.AppType

	return &m
}
