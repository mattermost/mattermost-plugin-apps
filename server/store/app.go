package store

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type AppID string

type Manifest struct {
	AppID                AppID       `json:"app_id"`
	OAuth2CallbackURL    string      `json:"oauth2_callback_url,omitempty"`
	Description          string      `json:"description,omitempty"`
	DisplayName          string      `json:"display_name,omitempty"`
	HomepageURL          string      `json:"homepage_url,omitempty"`
	Install              *Wish       `json:"install,omitempty"`
	RequestedPermissions Permissions `json:"requested_permissions,omitempty"`
	RootURL              string      `json:"root_url"`
}

type App struct {
	Manifest *Manifest `json:"manifest"`

	// Secret is used to issue JWT
	Secret string `json:",omitempty"`

	OAuth2ClientID string `json:",omitempty"`
	// Should secret be here? Or should we just fetch it using the ID?
	OAuth2ClientSecret string `json:",omitempty"`

	BotUserID      string `json:",omitempty"`
	BotUsername    string `json:",omitempty"`
	BotAccessToken string `json:",omitempty"`

	// Grants should be scopable in the future, per team, channel, post with regexp
	GrantedPermissions     Permissions `json:",omitempty"`
	NoUserConsentForOAuth2 bool        `json:",omitempty"`
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
