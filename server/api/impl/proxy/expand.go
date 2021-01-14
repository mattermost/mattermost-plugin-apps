package proxy

import (
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type expander struct {
	// Context to expand (can be expanded multiple times on the same expander)
	*api.Context

	mm           *pluginapi.Client
	conf         api.Configurator
	store        api.Store
	sessionToken api.SessionToken
}

func (p *Proxy) newExpander(cc *api.Context, mm *pluginapi.Client, conf api.Configurator, store api.Store, debugSessionToken api.SessionToken) *expander {
	e := &expander{
		Context:      cc,
		mm:           mm,
		conf:         conf,
		store:        store,
		sessionToken: debugSessionToken,
	}
	return e
}

func (e *expander) ExpandForApp(app *api.App, expand *api.Expand) (*api.Context, error) {
	clone := *e.Context
	clone.AppID = app.Manifest.AppID

	if e.MattermostSiteURL == "" {
		mmconf := e.conf.GetMattermostConfig()
		if mmconf.ServiceSettings.SiteURL != nil {
			e.MattermostSiteURL = *mmconf.ServiceSettings.SiteURL
		}
	}

	clone.MattermostSiteURL = e.MattermostSiteURL
	clone.BotUserID = app.BotUserID
	if expand == nil {
		clone.ExpandedContext = api.ExpandedContext{
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

	clone.ExpandedContext = api.ExpandedContext{
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

func stripUser(user *model.User, level api.ExpandLevel) *model.User {
	if user == nil || level == api.ExpandAll {
		return user
	}
	if level != api.ExpandSummary {
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

func stripChannel(channel *model.Channel, level api.ExpandLevel) *model.Channel {
	if channel == nil || level == api.ExpandAll {
		return channel
	}
	if level != api.ExpandSummary {
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

func stripTeam(team *model.Team, level api.ExpandLevel) *model.Team {
	if team == nil || level == api.ExpandAll {
		return team
	}
	if level != api.ExpandSummary {
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

func stripPost(post *model.Post, level api.ExpandLevel) *model.Post {
	if post == nil || level == api.ExpandAll {
		return post
	}
	if level != api.ExpandSummary {
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

func stripApp(app *api.App, level api.ExpandLevel) *api.App {
	if app == nil {
		return nil
	}

	clone := *app
	clone.Secret = ""
	clone.OAuth2ClientSecret = ""

	switch level {
	case api.ExpandAll, api.ExpandSummary:
		return &clone
	}
	return nil
}
