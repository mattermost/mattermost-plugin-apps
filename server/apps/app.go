package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type AppID string

type Manifest struct {
	AppID       AppID  `json:"app_id"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`

	OAuth2CallbackURL string `json:"oauth2_callback_url,omitempty"`
	HomepageURL       string `json:"homepage_url,omitempty"`
	RootURL           string `json:"root_url"`

	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`
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

	// Grants should be scopable in the future, per team, channel, post with regexp
	GrantedPermissions Permissions `json:"granted_permissions,omitempty"`
}

type PermissionType string

const (
	PermissionUserJoinedChannelNotification = PermissionType("user_joined_channel_notification")
	PermissionAddToPostMenu                 = PermissionType("add_to_post_menu")
	PermissionAddGrants                     = PermissionType("add_grants")
	PermissionActAsUser                     = PermissionType("act_as_user")
	PermissionActAsBot                      = PermissionType("act_as_bot")
)

func (p PermissionType) Markdown() md.MD {
	m := ""
	switch p {
	case PermissionAddToPostMenu:
		m = "Add items to Post menu"
	case PermissionUserJoinedChannelNotification:
		m = "Be notified when users join channels"
	case PermissionAddGrants:
		m = "Add more grants (WITHOUT ADDITIONAL ADMIN CONSENT)"
	case PermissionActAsUser:
		m = "Use Mattermost REST API as connected users"
	case PermissionActAsBot:
		m = "Use Mattermost REST API as the app's bot user"
	default:
		m = "unknown permission: " + string(p)
	}
	return md.MD(m)
}

type Permissions []PermissionType

func (p Permissions) Contains(permission PermissionType) bool {
	for _, current := range p {
		if current == permission {
			return true
		}
	}
	return false
}
