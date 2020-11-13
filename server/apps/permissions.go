package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Permissions []PermissionType

type PermissionType string

const (
	PermissionUserJoinedChannelNotification = PermissionType("user_joined_channel_notification")
	PermissionAddGrants                     = PermissionType("add_grants")
	PermissionActAsUser                     = PermissionType("act_as_user")
	PermissionActAsBot                      = PermissionType("act_as_bot")
)

func (p Permissions) Contains(permission PermissionType) bool {
	for _, current := range p {
		if current == permission {
			return true
		}
	}
	return false
}

func (p PermissionType) Markdown() md.MD {
	m := ""
	switch p {
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
