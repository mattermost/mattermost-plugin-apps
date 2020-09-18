package apps

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type AppID string

type PermissionType string

const (
	PermissionUserJoinedChannelNotification = PermissionType("user_joined_channel_notification")
	PermissionAddToPostMenu                 = PermissionType("add_to_post_menu")
	PermissionAddGrants                     = PermissionType("add_grants")
	PermissionActAsUser                     = PermissionType("act_as_user")
	PermissionActAsBot                      = PermissionType("act_as_bot")
)

func (p PermissionType) String() string {
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
		m = fmt.Sprintf("Use Mattermost REST API as @%s bot", constants.BotUserName)
	default:
		m = "unknown permission: " + string(p)
	}
	return m
}

func (p PermissionType) Markdown() md.MD {
	return md.MD(p.String())
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

type Manifest struct {
	AppID                AppID
	DisplayName          string
	Description          string
	RootURL              string
	RequestedPermissions Permissions
}

type App struct {
	Manifest *Manifest

	// Secret is used to issue JWT
	Secret string

	OAuthAppID string
	// Should secret be here? Or should we just fetch it using the ID?
	OAuthSecret string

	BotID    string
	BotToken string
	// Grants should be scopable in the future, per team, channel, post with regexp
	GrantedPermissions     Permissions
	NoUserConsentForOAuth2 bool
}
