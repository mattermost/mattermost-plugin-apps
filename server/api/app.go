package api

type AppID string

type Manifest struct {
	AppID       AppID  `json:"app_id"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	OAuth2CallbackURL string `json:"oauth2_callback_url,omitempty"`
	HomepageURL       string `json:"homepage_url,omitempty"`
	RootURL           string `json:"root_url"`

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`

	// RequestedLocations is the list of top-level locations that the
	// application intends to bind to, e.g. `{"/post_menu", "/channel_header",
	// "/command/apptrigger"}``.
	RequestedLocations Locations `json:"requested_locations,omitempty"`
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

func (m *Manifest) ConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"app_id":                string(m.AppID),
		"display_name":          m.DisplayName,
		"description":           m.Description,
		"oauth2_callback_url":   m.OAuth2CallbackURL,
		"homepage_url":          m.HomepageURL,
		"root_url":              m.RootURL,
		"requested_permissions": m.RequestedPermissions.toStringArray(),
		"requested_locations":   m.RequestedLocations.toStringArray(),
	}
}

func ManifestFromConfigMap(in interface{}) *Manifest {
	m := &Manifest{}
	c, _ := in.(map[string]interface{})
	if len(c) == 0 {
		return m
	}
	appID, _ := c["app_id"].(string)
	m.AppID = AppID(appID)
	m.DisplayName, _ = c["display_name"].(string)
	m.Description, _ = c["description"].(string)
	m.OAuth2CallbackURL = c["oauth2_callback_url"].(string)
	m.HomepageURL = c["homepage_url"].(string)
	m.RootURL = c["root_url"].(string)
	m.RequestedPermissions = permissionsFromConfigArray(c["requested_permissions"])
	m.RequestedLocations = locationsFromConfigArray(c["requested_locations"])
	return m
}

func (a *App) ConfigMap() map[string]interface{} {
	return map[string]interface{}{
		"manifest":             a.Manifest.ConfigMap(),
		"secret":               a.Secret,
		"oauth2_client_id":     a.OAuth2ClientID,
		"oauth2_client_secret": a.OAuth2ClientSecret,
		"oauth2_trusted_app":   a.OAuth2TrustedApp,
		"bot_user_id":          a.BotUserID,
		"bot_username":         a.BotUsername,
		"bot_access_token":     a.BotAccessToken,
		"granted_permissions":  a.GrantedPermissions.toStringArray(),
		"granted_locations":    a.GrantedLocations.toStringArray(),
	}
}

func AppFromConfigMap(in interface{}) *App {
	c, _ := in.(map[string]interface{})
	if len(c) == 0 {
		return &App{}
	}

	a := &App{}
	a.Manifest = ManifestFromConfigMap(c["manifest"])
	a.Secret, _ = c["secret"].(string)
	a.OAuth2ClientID, _ = c["oauth2_client_id"].(string)
	a.OAuth2ClientSecret, _ = c["oauth2_client_secret"].(string)
	a.OAuth2TrustedApp, _ = c["oauth2_trusted_app"].(bool)
	a.BotUserID, _ = c["bot_user_id"].(string)
	a.BotUsername, _ = c["bot_username"].(string)
	a.BotAccessToken, _ = c["bot_access_token"].(string)
	a.GrantedPermissions = permissionsFromConfigArray(c["granted_permissions"])
	a.GrantedLocations = locationsFromConfigArray(c["granted_locations"])
	return a
}
