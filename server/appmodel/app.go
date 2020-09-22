// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package appmodel

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
