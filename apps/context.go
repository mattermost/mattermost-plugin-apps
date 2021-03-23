package apps

import (
	"github.com/mattermost/mattermost-server/v5/model"
)

// Context is included in CallRequest and provides App with information about
// Mattermost environment (configuration, authentication), and the context of
// the user agent (current channel, etc.)
//
// To help reduce the need to go back to Mattermost REST API, ExpandedContext
// can be included by adding a corresponding Expand attribute to the originating
// Call.
type Context struct {
	// AppID is used for handling CallRequest internally.
	AppID AppID `json:"app_id"`

	// Fully qualified original Location of the user action (if applicable),
	// e.g. "/command/helloworld/send" or "/channel_header/send".
	Location Location `json:"location,omitempty"`

	// Subject is a subject of notification, if the call originated from a
	// subscription.
	Subject Subject `json:"subject,omitempty"`

	// BotUserID of the App.
	BotUserID string `json:"bot_user_id,omitempty"`

	// ActingUserID is primarily (or exclusively?) for calls originating from
	// user submissions.
	ActingUserID string `json:"acting_user_id,omitempty"`

	// UserID indicates the subject of the command. Once Mentions is
	// implemented, it may be replaced by Mentions.
	UserID string `json:"user_id,omitempty"`

	// The optional IDs of Mattermost entities associated with the call: Team,
	// Channel, Post, RootPost.
	TeamID     string `json:"team_id"`
	ChannelID  string `json:"channel_id,omitempty"`
	PostID     string `json:"post_id,omitempty"`
	RootPostID string `json:"root_post_id,omitempty"`

	// Top-level Mattermost site URL to use for REST API calls.
	MattermostSiteURL string `json:"mattermost_site_url"`

	// App's path on the Mattermost instance (appendable to MattermostSiteURL).
	AppPath string `json:"app_path"`

	// UserAgent used to perform the call. It can be either "webapp" or "mobile".
	// Non user interactions like notifications will have this field empty.
	UserAgent string `json:"user_agent,omitempty"`

	// More data as requested by call.Expand
	ExpandedContext
}

// ExpandedContext contains authentication, and Mattermost entity data, as
// indicated by the Expand attribute of the originating Call.
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

func WithActingUser(id string) func(c Context) *Context {
	return func(c Context) *Context {
		c.ActingUserID = id
		return &c
	}
}

func WithTeam(id string) func(c Context) *Context {
	return func(c Context) *Context {
		c.TeamID = id
		return &c
	}
}

func ForChannelCreated(ch *model.Channel) func(c Context) *Context {
	return func(c Context) *Context {
		c.UserID = ch.CreatorId
		c.ChannelID = ch.Id
		c.TeamID = ch.TeamId
		c.ExpandedContext.Channel = ch
		return &c
	}
}

func ForPostCreated(p *model.Post) func(c Context) *Context {
	return func(c Context) *Context {
		c.UserID = p.UserId
		c.PostID = p.Id
		c.RootPostID = p.RootId
		c.ChannelID = p.ChannelId
		c.ExpandedContext.Post = p
		return &c
	}
}

func ForUserCreated(user *model.User) func(c Context) *Context {
	return func(c Context) *Context {
		c.UserID = user.Id
		c.ExpandedContext.User = user
		return &c
	}
}

func ForTeamMember(tm *model.TeamMember, actingUser *model.User) func(c Context) *Context {
	return func(c Context) *Context {
		actingUserID := ""
		if actingUser != nil {
			actingUserID = actingUser.Id
		}
		c.ActingUserID = actingUserID
		c.UserID = tm.UserId
		c.TeamID = tm.TeamId
		c.ExpandedContext.ActingUser = actingUser
		return &c
	}
}

func ForChannelMember(cm *model.ChannelMember, actingUser *model.User) func(c Context) *Context {
	return func(c Context) *Context {
		actingUserID := ""
		if actingUser != nil {
			actingUserID = actingUser.Id
		}
		c.ActingUserID = actingUserID
		c.UserID = cm.UserId
		c.ChannelID = cm.ChannelId
		c.ExpandedContext.ActingUser = actingUser
		return &c
	}
}

func ForCommand(commandArgs *model.CommandArgs) func(c Context) *Context {
	return func(c Context) *Context {
		c.ActingUserID = commandArgs.UserId
		c.UserID = commandArgs.UserId
		c.TeamID = commandArgs.TeamId
		c.ChannelID = commandArgs.ChannelId
		c.MattermostSiteURL = commandArgs.SiteURL
		return &c
	}
}
