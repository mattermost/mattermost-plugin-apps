package api

import "encoding/json"

type AppID string

type Manifest struct {
	AppID       AppID  `json:"app_id"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	HomepageURL string `json:"homepage_url,omitempty"`

	RootURL                 string `json:"root_url,omitempty"`
	RemoteOAuth2CallbackURL string `json:"remote_oauth2_callback_url,omitempty"`

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`

	// By default invoke "/install", expanding App, AdminAccessToken, and Config
	Install *Call `json:"install,omitempty"`

	// Deployment manifest for hostable apps will include path->invoke mappings
}

var DefaultInstall = &Call{
	URL: AppInstallPath,
	Expand: &Expand{
		App:              ExpandAll,
		AdminAccessToken: ExpandAll,
		Config:           ExpandAll,
	},
}

type App struct {
	Manifest *Manifest `json:"manifest"`

	// Secret is used to issue JWT
	Secret string `json:"secret,omitempty"`

	OAuth2ClientID     string `json:"oauth2_client_id,omitempty"`
	OAuth2ClientSecret string `json:"oauth2_client_secret,omitempty"`
	OAuth2TrustedApp   bool   `json:"oauth2_trusted_app,omitempty"`

	BotUserID      string `json:"bot_user_id,omitempty"`
	BotUsername    string `json:"bot_username,omitempty"`
	BotAccessToken string `json:"bot_access_token,omitempty"`

	// Grants should be scopable in the future, per team, channel, post with
	// regexp.
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`

	// GrantedLocations contains the list of top locations that the
	// application is allowed to bind to.
	GrantedLocations Locations `json:"granted_locations,omitempty"`
}

func (a *App) ConfigMap() map[string]interface{} {
	b, _ := json.Marshal(a)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out
}

func AppFromConfigMap(in interface{}) *App {
	b, _ := json.Marshal(in)
	var out App
	_ = json.Unmarshal(b, &out)
	return &out
}
