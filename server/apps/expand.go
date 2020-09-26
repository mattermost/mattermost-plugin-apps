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
	ActingUser ExpandLevel `json:"acting_user"`
	Channel    ExpandLevel `json:"channel,omitempty"`
	Config     bool        `json:"config,omitempty"`
	Mentioned  ExpandLevel `json:"mentioned,omitempty"`
	ParentPost ExpandLevel `json:"parent_post,omitempty"`
	Post       ExpandLevel `json:"post,omitempty"`
	RootPost   ExpandLevel `json:"root_post,omitempty"`
	Team       ExpandLevel `json:"team,omitempty"`
	User       ExpandLevel `json:"user,omitempty"`
}

type Expanded struct {
	ActingUser *model.User       `json:"acting_user"`
	App        *App              `json:"app,omitempty"`
	Channel    *model.Channel    `json:"channel,omitempty"`
	Config     *MattermostConfig `json:"config,omitempty"`
	Mentioned  []*model.User     `json:"mentioned,omitempty"`
	ParentPost *model.Post       `json:"parent_post,omitempty"`
	Post       *model.Post       `json:"post,omitempty"`
	RootPost   *model.Post       `json:"root_post,omitempty"`
	Team       *model.Team       `json:"team,omitempty"`
	User       *model.User       `json:"user,omitempty"`
}

type MattermostConfig struct {
	SiteURL string `json:"site_url"`
}

type expander struct {
	mm           *pluginapi.Client
	configurator configurator.Service

	ActingUser *model.User
	Channel    *model.Channel
	Config     *model.Config
	User       *model.User
}

func NewExpander(mm *pluginapi.Client, configurator configurator.Service) Expander {
	return &expander{
		mm:           mm,
		configurator: configurator,
	}
}

func (e *expander) Expand(expand *Expand, actingUserID, userID, channelID string) (expanded *Expanded, err error) {
	for _, f := range []func(*Expand) error{
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

func (e *expander) collectConfig(expand *Expand) error {
	if e.Config != nil || !expand.Config {
		return nil
	}
	e.Config = e.configurator.GetMattermostConfig()
	return nil
}

func (e *expander) collectChannel(channelID string) func(*Expand) error {
	return func(expand *Expand) error {
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

func (e *expander) collectUser(userID string, userref **model.User) func(*Expand) error {
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

func (e *expander) produce(expand *Expand) *Expanded {
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
