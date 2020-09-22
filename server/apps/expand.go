// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/appmodel"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-server/v5/model"
)

type expander struct {
	mm           *pluginapi.Client
	configurator configurator.Service

	ActingUser *model.User
	Channel    *model.Channel
	Config     *model.Config
	User       *model.User
}

func NewExpander(mm *pluginapi.Client, configurator configurator.Service) appmodel.Expander {
	return &expander{
		mm:           mm,
		configurator: configurator,
	}
}

func (e *expander) Expand(expand *appmodel.Expand, actingUserID, userID, channelID string) (expanded *appmodel.Expanded, err error) {
	for _, f := range []func(*appmodel.Expand) error{
		e.collectConfig,
		e.collectUser(userID, &e.User),
		e.collectUser(actingUserID, &e.ActingUser),
		e.collectChannel(channelID),
	} {
		err = f(expand)
		if err != nil {
			return nil, err
		}
	}

	expanded = e.produce(expand)
	return expanded, nil
}

func (e *expander) collectConfig(expand *appmodel.Expand) error {
	if e.Config != nil || !expand.Config {
		return nil
	}
	e.Config = e.configurator.GetMattermostConfig()
	return nil
}

func (e *expander) collectChannel(channelID string) func(*appmodel.Expand) error {
	return func(expand *appmodel.Expand) error {
		if channelID == "" || !isValidExpandLevel(expand.Channel) {
			return nil
		}

		mmchannel, err := e.mm.Channel.Get(channelID)
		if err != nil {
			return err
		}

		e.Channel = mmchannel
		return nil
	}
}

func (e *expander) collectUser(userID string, userref **model.User) func(*appmodel.Expand) error {
	return func(expand *appmodel.Expand) error {
		if *userref != nil || userID == "" || !isValidExpandLevel(expand.User) {
			return nil
		}

		mmuser, err := e.mm.User.Get(userID)
		if err != nil {
			return err
		}
		mmuser.SanitizeProfile(nil)

		*userref = mmuser
		return nil
	}
}

func (e *expander) produce(expand *appmodel.Expand) *appmodel.Expanded {
	expanded := &appmodel.Expanded{}

	if expand.Config {
		expanded.Config = &appmodel.MattermostConfig{}
		if e.Config.ServiceSettings.SiteURL != nil {
			expanded.Config.SiteURL = *e.Config.ServiceSettings.SiteURL
		}
	}

	expanded.User = produceUser(e.User, expand)
	expanded.ActingUser = produceUser(e.ActingUser, expand)
	return nil
}

func produceUser(user *model.User, expand *appmodel.Expand) *model.User {
	if expand.User == "" || !isValidExpandLevel(expand.User) {
		return nil
	}

	switch expand.User {
	case appmodel.ExpandSummary:
		return &model.User{
			Id:             user.Id,
			Username:       user.Username,
			Email:          user.Email,
			Nickname:       user.Nickname,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Roles:          user.Roles,
			Locale:         user.Locale,
			Timezone:       user.Timezone,
			IsBot:          user.IsBot,
			BotDescription: user.BotDescription,
		}

	case appmodel.ExpandAll:
		return user
	}

	return nil
}

func isValidExpandLevel(l appmodel.ExpandLevel) bool {
	return l == appmodel.ExpandAll || l == appmodel.ExpandSummary
}
