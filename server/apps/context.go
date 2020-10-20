package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Context struct {
	AppID        store.AppID       `json:"app_id"`
	ActingUserID string            `json:"acting_user_id,omitempty"`
	UserID       string            `json:"user_id,omitempty"`
	TeamID       string            `json:"team_id"`
	ChannelID    string            `json:"channel_id,omitempty"`
	PostID       string            `json:"post_id,omitempty"`
	RootPostID   string            `json:"root_post_id,omitempty"`
	Props        map[string]string `json:"props,omitempty"`
	expandedContext
}

type expandedContext struct {
	ActingUser *model.User       `json:"acting_user,omitempty"`
	App        *store.App        `json:"app,omitempty"`
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

func NewChannelContext(ch *model.Channel) *Context {
	return &Context{
		UserID:    ch.CreatorId,
		ChannelID: ch.Id,
		TeamID:    ch.TeamId,
		expandedContext: expandedContext{
			Channel: ch,
		},
	}
}

func NewPostContext(p *model.Post) *Context {
	return &Context{
		UserID:     p.UserId,
		PostID:     p.Id,
		RootPostID: p.RootId,
		ChannelID:  p.ChannelId,
		expandedContext: expandedContext{
			Post: p,
		},
	}
}

func NewUserContext(user *model.User) *Context {
	return &Context{
		UserID: user.Id,
		expandedContext: expandedContext{
			User: user,
		},
	}
}

func NewTeamMemberContext(tm *model.TeamMember, actingUser *model.User) *Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return &Context{
		ActingUserID: actingUserID,
		UserID:       tm.UserId,
		TeamID:       tm.TeamId,
		expandedContext: expandedContext{
			ActingUser: actingUser,
		},
	}
}

func NewChannelMemberContext(cm *model.ChannelMember, actingUser *model.User) *Context {
	actingUserID := ""
	if actingUser != nil {
		actingUserID = actingUser.Id
	}
	return &Context{
		ActingUserID: actingUserID,
		UserID:       cm.UserId,
		ChannelID:    cm.ChannelId,
		expandedContext: expandedContext{
			ActingUser: actingUser,
		},
	}
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
