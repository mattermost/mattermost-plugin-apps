package proxy

import (
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

type expander struct {
	// Context to expand (can be expanded multiple times on the same expander)
	*apps.Context

	mm           *pluginapi.Client
	conf         config.Service
	store        *store.Service
	sessionToken apps.SessionToken
}

func (p *Proxy) newExpander(cc *apps.Context, mm *pluginapi.Client, conf config.Service, store *store.Service, debugSessionToken apps.SessionToken) *expander {
	e := &expander{
		Context:      cc,
		mm:           mm,
		conf:         conf,
		store:        store,
		sessionToken: debugSessionToken,
	}
	return e
}

func (e *expander) ExpandForApp(app *apps.App, expand *apps.Expand) (*apps.Context, error) {
	clone := *e.Context
	clone.AppID = app.AppID

	if e.MattermostSiteURL == "" {
		mmconf := e.conf.GetMattermostConfig()
		if mmconf.ServiceSettings.SiteURL != nil {
			e.MattermostSiteURL = *mmconf.ServiceSettings.SiteURL
		}
	}

	clone.MattermostSiteURL = e.MattermostSiteURL
	clone.BotUserID = app.BotUserID
	if expand == nil {
		clone.ExpandedContext = apps.ExpandedContext{
			BotAccessToken: app.BotAccessToken,
		}
		return &clone, nil
	}

	if expand.ActingUser != "" && e.ActingUserID != "" && e.ActingUser == nil {
		actingUser, err := e.mm.User.Get(e.ActingUserID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand acting user %s", e.ActingUserID)
		}
		e.ActingUser = actingUser
	}

	if expand.Channel != "" && e.ChannelID != "" && e.Channel == nil {
		ch, err := e.mm.Channel.Get(e.ChannelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand channel %s", e.ChannelID)
		}
		e.Channel = ch
	}

	// TODO expand Mentioned

	if expand.Post != "" && e.PostID != "" && e.Post == nil {
		post, err := e.mm.Post.GetPost(e.PostID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand post %s", e.PostID)
		}
		e.Post = post
	}

	if expand.RootPost != "" && e.RootPostID != "" && e.RootPost == nil {
		post, err := e.mm.Post.GetPost(e.RootPostID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand root post %s", e.RootPostID)
		}
		e.RootPost = post
	}

	if expand.Team != "" && e.TeamID != "" && e.Team == nil {
		team, err := e.mm.Team.Get(e.TeamID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand team %s", e.TeamID)
		}
		e.Team = team
	}

	if expand.User != "" && e.UserID != "" && e.User == nil {
		user, err := e.mm.User.Get(e.UserID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand user %s", e.UserID)
		}
		e.User = user
	}

	clone.ExpandedContext = apps.ExpandedContext{
		BotAccessToken: app.BotAccessToken,

		ActingUser: stripUser(e.ActingUser, expand.ActingUser),
		App:        stripApp(app, expand.App),
		Channel:    stripChannel(e.Channel, expand.Channel),
		Post:       stripPost(e.Post, expand.Post),
		RootPost:   stripPost(e.RootPost, expand.RootPost),
		Team:       stripTeam(e.Team, expand.Team),
		User:       stripUser(e.User, expand.User),
		// TODO Mentioned
	}

	// TODO: use the appropriate user's OAuth2 token once re-implemented, for
	// now pass in the session token to make things work.
	if expand.AdminAccessToken != "" {
		clone.ExpandedContext.AdminAccessToken = string(e.sessionToken)
	}
	if expand.ActingUserAccessToken != "" {
		clone.ExpandedContext.ActingUserAccessToken = string(e.sessionToken)
	}

	return &clone, nil
}

func stripUser(user *model.User, level apps.ExpandLevel) *model.User {
	if user == nil {
		return user
	}
	if level == apps.ExpandAll {
		sanitized := *user
		sanitized.Sanitize(map[string]bool{
			"passwordupdate": true,
			"authservice":    true,
		})
		return &sanitized
	}
	if level != apps.ExpandSummary {
		return nil
	}
	return &model.User{
		BotDescription: user.BotDescription,
		DeleteAt:       user.DeleteAt,
		Email:          user.Email,
		FirstName:      user.FirstName,
		Id:             user.Id,
		IsBot:          user.IsBot,
		LastName:       user.LastName,
		Locale:         user.Locale,
		Nickname:       user.Nickname,
		Roles:          user.Roles,
		Timezone:       user.Timezone,
		Username:       user.Username,
	}
}

func stripChannel(channel *model.Channel, level apps.ExpandLevel) *model.Channel {
	if channel == nil {
		return channel
	}
	if level == apps.ExpandAll {
		return channel
	}
	if level != apps.ExpandSummary {
		return nil
	}
	return &model.Channel{
		Id:          channel.Id,
		DeleteAt:    channel.DeleteAt,
		TeamId:      channel.TeamId,
		Type:        channel.Type,
		DisplayName: channel.DisplayName,
		Name:        channel.Name,
	}
}

func stripTeam(team *model.Team, level apps.ExpandLevel) *model.Team {
	if team == nil {
		return team
	}
	if level == apps.ExpandAll {
		sanitized := *team
		sanitized.Sanitize()
		return &sanitized
	}
	if level != apps.ExpandSummary {
		return nil
	}
	return &model.Team{
		Id:          team.Id,
		DisplayName: team.DisplayName,
		Name:        team.Name,
		Description: team.Description,
		Email:       team.Email,
		Type:        team.Type,
	}
}

func stripPost(post *model.Post, level apps.ExpandLevel) *model.Post {
	if post == nil {
		return post
	}
	if level == apps.ExpandAll {
		sanitized := *post.Clone()
		sanitized.SanitizeProps()
		return &sanitized
	}
	if level != apps.ExpandSummary {
		return nil
	}
	return &model.Post{
		Id:        post.Id,
		Type:      post.Type,
		UserId:    post.UserId,
		ChannelId: post.ChannelId,
		RootId:    post.RootId,
		Message:   post.Message,
	}
}

func stripApp(app *apps.App, level apps.ExpandLevel) *apps.App {
	if app == nil {
		return nil
	}

	clone := *app
	clone.Secret = ""
	clone.OAuth2ClientSecret = ""

	switch level {
	case apps.ExpandAll, apps.ExpandSummary:
		return &clone
	}
	return nil
}
