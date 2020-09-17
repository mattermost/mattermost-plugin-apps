// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-plugin-apps/server/configurator"
	"github.com/mattermost/mattermost-server/v5/model"
)

type ExpandEntity string

const (
	ExpandActingUser = ExpandEntity("ActingUser")
	ExpandUser       = ExpandEntity("User")
	ExpandChannel    = ExpandEntity("Channel")
	ExpandConfig     = ExpandEntity("Config")
)

type ExpandLevel string

const (
	ExpandAll     = ExpandLevel("All")
	ExpandSummary = ExpandLevel("Summary")
)

type Expand struct {
	ActingUser ExpandLevel
	Channel    ExpandLevel
	Config     bool
	User       ExpandLevel
	Post       ExpandLevel
	ParentPost ExpandLevel
	RootPost   ExpandLevel
	Team       ExpandLevel
	Mentioned  ExpandLevel
}

type Expanded struct {
	ActingUser *model.User
	Channel    *model.Channel
	Config     *MattermostConfig
	User       *model.User
	Post       *model.Post
	ParentPost *model.Post
	RootPost   *model.Post
	Team       *model.Team
	Mentioned  []*model.User
}

type MattermostConfig struct {
	SiteURL string
}

type Expander struct {
	mm           *pluginapi.Client
	configurator configurator.Configurator

	ActingUser *model.User
	Team       *model.Team
	Channel    *model.Channel
	Config     *model.Config
	User       *model.User
}

func NewExpander(mm *pluginapi.Client, configurator configurator.Configurator) *Expander {
	return &Expander{
		mm:           mm,
		configurator: configurator,
	}
}

func (e *Expander) Expand(expand *Expand, actingUserID, userID, channelID string) (expanded *Expanded, err error) {
	for _, f := range []func(*Expand) error{
		e.collectConfig,
		e.collectUser(userID, &e.User),
		e.collectUser(actingUserID, &e.ActingUser),
		e.collectChannelAndTeam(channelID),
	} {
		err = f(expand)
		if err != nil {
			return nil, err
		}
	}

	expanded = e.produce(expand)
	return expanded, nil
}

func (e *Expander) collectConfig(expand *Expand) error {
	if e.Config != nil || !expand.Config {
		return nil
	}
	e.Config = e.configurator.GetMattermostConfig()
	return nil
}

func (e *Expander) collectChannelAndTeam(channelID string) func(*Expand) error {
	return func(expand *Expand) error {
		if channelID == "" || !isValidExpandLevel(expand.Channel) {
			return nil
		}

		mmchannel, err := e.mm.Channel.Get(channelID)
		if err != nil {
			return err
		}

		mmteam, err := e.mm.Team.Get(mmchannel.TeamId)
		if err != nil {
			return err
		}

		e.Channel = mmchannel
		if e.Team == nil {
			e.Team = mmteam
		}
		return nil
	}
}

func (e *Expander) collectUser(userID string, userref **model.User) func(*Expand) error {
	return func(expand *Expand) error {
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

func (e *Expander) produce(expand *Expand) *Expanded {
	expanded := &Expanded{}

	if expand.Config {
		expanded.Config = &MattermostConfig{}
		if e.Config.ServiceSettings.SiteURL != nil {
			expanded.Config.SiteURL = *e.Config.ServiceSettings.SiteURL
		}
	}

	expanded.User = produceUser(e.User, expand)
	expanded.ActingUser = produceUser(e.ActingUser, expand)
	return nil
}

func produceUser(user *model.User, expand *Expand) *model.User {
	if expand.User == "" || !isValidExpandLevel(expand.User) {
		return nil
	}

	switch expand.User {
	case ExpandSummary:
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

	case ExpandAll:
		return user
	}

	return nil
}

func isValidExpandLevel(l ExpandLevel) bool {
	return l == ExpandAll || l == ExpandSummary
}
