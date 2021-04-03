package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Permissions []Permission

type Permission string

const (
	// PermissionUserJoinedChannelNotification means that the app is allowed to
	// receive user_joined_channel notifications
	PermissionUserJoinedChannelNotification Permission = "user_joined_channel_notification"

	// PermissionActAsBot means that a Bot User will be created when the App is
	// installed. Call requests will then include the Bot access token, the app
	// can use them with the Mattermost REST API. The bot will not automatically
	// receive permissions to any resources, need to be added explicitly.
	PermissionActAsBot Permission = "act_as_bot"

	// PermissionActAsUser means that the app is allowed to connect users'
	// OAuth2 accounts, and then use user API tokens.
	PermissionActAsUser Permission = "act_as_user"

	// PermissionActAsAdmin means that the app is allowed to request admin-level
	// access tokens in its calls.
	PermissionActAsAdmin Permission = "act_as_admin"

	// PermissionRemoteOAuth2 means that the app is allowed to use remote (3rd
	// party) OAuth2 support, and will store secrets to 3rd party system(s).
	PermissionRemoteOAuth2 Permission = "remote_oauth2"
)

func (p Permissions) Contains(permission Permission) bool {
	for _, current := range p {
		if current == permission {
			return true
		}
	}
	return false
}

func (p Permission) Markdown() md.MD {
	m := ""
	switch p {
	case PermissionUserJoinedChannelNotification:
		m = "Be notified when users join channels"
	case PermissionActAsAdmin:
		m = "Use Mattermost REST API as a sysadmin"
	case PermissionActAsUser:
		m = "Use Mattermost REST API as connected users"
	case PermissionActAsBot:
		m = "Use Mattermost REST API as the app's bot user"
	case PermissionRemoteOAuth2:
		m = "Use remote (3rd party) OAuth2 and store secrets"
	default:
		m = "unknown permission: " + string(p)
	}
	return md.MD(m)
}

func (p Permissions) IsValid() error {
	// Check for permission dependencies. (P1, P2, ..., PN) means P1 requires
	// (depends on) P2...PN.
	for _, pp := range []Permissions{
		{PermissionRemoteOAuth2, PermissionActAsUser, PermissionActAsAdmin},
		{PermissionUserJoinedChannelNotification, PermissionActAsBot},
	} {
		if len(pp) == 0 || !p.Contains(pp[0]) {
			continue
		}
		for _, d := range pp[1:] {
			if !p.Contains(d) {
				return utils.NewInvalidError("%s requires %s", p, d)
			}
		}
	}

	return nil
}
