// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/constants"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

type SessionToken string

type API interface {
	Call(Call) (*CallResponse, error)
	InstallApp(*InInstallApp, *CallContext, SessionToken) (*App, md.MD, error)
	ProvisionApp(*InProvisionApp, *CallContext, SessionToken) (*App, md.MD, error)
	Notify(
		subject constants.SubscriptionSubject,
		tm *model.TeamMember,
		cm *model.ChannelMember,
		actingUser *model.User,
		channel *model.Channel,
		post *model.Post,
	) error
}
