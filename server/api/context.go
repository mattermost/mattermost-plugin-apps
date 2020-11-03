package api

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type Context struct {
	AppID        AppID             `json:"app_id"`
	LocationID   LocationID        `json:"location_id,omitempty"`
	ActingUserID string            `json:"acting_user_id,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	TeamID       string            `json:"team_id"`
	ChannelID    string            `json:"channel_id,omitempty"`
	PostID       string            `json:"post_id,omitempty"`
	RootPostID   string            `json:"root_post_id,omitempty"`
	Props        map[string]string `json:"props,omitempty"`
	ExpandedContext
}

type ExpandedContext struct {
	ActingUser *model.User       `json:"acting_user,omitempty"`
	App        *App              `json:"app,omitempty"`
	Channel    *model.Channel    `json:"channel,omitempty"`
	Config     *MattermostConfig `json:"config,omitempty"`
	Mentioned  []*model.User     `json:"mentioned,omitempty"`
	Post       *model.Post       `json:"post,omitempty"`
	RootPost   *model.Post       `json:"root_post,omitempty"`
	Team       *model.Team       `json:"team,omitempty"`
	User       *model.User       `json:"user,omitempty"`
}

type MattermostConfig struct {
	SiteURL string `json:"site_url"`
}

type Thread struct {
	ChannelID  string `json:"channel_id"`
	RootPostID string `json:"root_post_id"`
}

func (cc *Context) GetProp(n string) string {
	if len(cc.Props) == 0 {
		return ""
	}
	return cc.Props[n]
}

func (cc *Context) SetProp(n, v string) {
	if len(cc.Props) == 0 {
		cc.Props = map[string]string{}
	}
	cc.Props[n] = v
}
