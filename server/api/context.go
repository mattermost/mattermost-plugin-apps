package api

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

type Context struct {
	AppID             AppID             `json:"app_id"`
	Location          Location          `json:"location,omitempty"`
	Subject           Subject           `json:"subject,omitempty"`
	BotUserID         string            `json:"bot_user_id,omitempty"`
	ActingUserID      string            `json:"acting_user_id,omitempty"`
	UserID            string            `json:"user_id,omitempty"`
	TeamID            string            `json:"team_id"`
	ChannelID         string            `json:"channel_id,omitempty"`
	PostID            string            `json:"post_id,omitempty"`
	RootPostID        string            `json:"root_post_id,omitempty"`
	Props             map[string]string `json:"props,omitempty"`
	MattermostSiteURL string            `json:"mattermost_site_url"`
	ExpandedContext
}

type ExpandedContext struct {
	//  BotAccessToken is always provided in expanded context
	BotAccessToken string `json:"bot_access_token,omitempty"`

	ActingUser            *model.User    `json:"acting_user,omitempty"`
	ActingUserAccessToken string         `json:"acting_user_access_token,omitempty"`
	AdminAccessToken      string         `json:"admin_access_token,omitempty"`
	App                   *App           `json:"app,omitempty"`
	Channel               *model.Channel `json:"channel,omitempty"`
	Mentioned             []*model.User  `json:"mentioned,omitempty"`
	Post                  *model.Post    `json:"post,omitempty"`
	RootPost              *model.Post    `json:"root_post,omitempty"`
	Team                  *model.Team    `json:"team,omitempty"`

	// TODO replace User with mentions
	User *model.User `json:"user,omitempty"`
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

func NewChannelContext(ch *model.Channel) *Context {
	return &Context{
		UserID:    ch.CreatorId,
		ChannelID: ch.Id,
		TeamID:    ch.TeamId,
		ExpandedContext: ExpandedContext{
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
		ExpandedContext: ExpandedContext{
			Post: p,
		},
	}
}

func NewUserContext(user *model.User) *Context {
	return &Context{
		UserID: user.Id,
		ExpandedContext: ExpandedContext{
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
		ExpandedContext: ExpandedContext{
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
		ExpandedContext: ExpandedContext{
			ActingUser: actingUser,
		},
	}
}

func NewCommandContext(commandArgs *model.CommandArgs) *Context {
	return &Context{
		ActingUserID: commandArgs.UserId,
		UserID:       commandArgs.UserId,
		TeamID:       commandArgs.TeamId,
		ChannelID:    commandArgs.ChannelId,
	}
}
