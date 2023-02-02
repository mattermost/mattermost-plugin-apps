// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type Permissions []Permission

func (list Permissions) String() string {
	m := ""
	for i, permission := range list {
		if i != 0 {
			m += ", "
		}
		m += permission.String()
	}

	return m
}

func (list Permissions) Contains(permission Permission) bool {
	for _, current := range list {
		if current == permission {
			return true
		}
	}
	return false
}

type Permission string

const (
	// PermissionActAsBot means that a Bot User will be created when the App is
	// installed. Call requests will then include the Bot access token, the app
	// can use them with the Mattermost REST API. The bot will not automatically
	// receive permissions to any resources, need to be added explicitly.
	PermissionActAsBot Permission = "act_as_bot"

	// PermissionActAsUser means that the app is allowed to connect users'
	// OAuth2 accounts, and then use user API tokens.
	PermissionActAsUser Permission = "act_as_user"

	// PermissionRemoteOAuth2 means that the app is allowed to use remote (3rd
	// party) OAuth2 support, and will store secrets to 3rd party system(s).
	PermissionRemoteOAuth2 Permission = "remote_oauth2"

	// PermissionRemoteWebhooks means that the app is allowed to receive webhooks from a remote (3rd
	// party) system, and process them as Bot.
	PermissionRemoteWebhooks Permission = "remote_webhooks"
)

func (p Permission) String() string {
	m := ""
	switch p {
	case PermissionActAsUser:
		m = "Use Mattermost REST API as the connected user"
	case PermissionActAsBot:
		m = "Use Mattermost REST API as the app's bot user"
	case PermissionRemoteOAuth2:
		m = "use a remote (3rd party) OAuth2 and store secrets"
	case PermissionRemoteWebhooks:
		m = "receive webhook messages from a remote (3rd party) system"
	default:
		m = "unknown permission: " + string(p)
	}
	return m
}

func (list Permissions) Validate() error {
	if len(list) == 0 {
		return nil
	}
	// Check for permission dependencies. (P1, P2, ..., PN) means P1 requires
	// (depends on) P2...PN.
	for _, p := range []Permissions{
		{PermissionRemoteWebhooks, PermissionActAsBot},
		{PermissionRemoteOAuth2, PermissionActAsUser},
	} {
		if len(p) == 0 || !list.Contains(p[0]) {
			continue
		}
		for _, d := range p[1:] {
			if !list.Contains(d) {
				return utils.NewInvalidError("%s requires %s", list, d)
			}
		}
	}

	return nil
}
