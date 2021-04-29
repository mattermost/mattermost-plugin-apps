package proxy

import (
	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils"
)

type expander struct {
	// Context to expand (can be expanded multiple times on the same expander)
	*apps.Context

	mm        *pluginapi.Client
	conf      config.Service
	store     *store.Service
	sessionID string
	session   *model.Session
}

func (p *Proxy) newExpander(cc *apps.Context, mm *pluginapi.Client, conf config.Service, store *store.Service, sessionID string) *expander {
	e := &expander{
		Context:   cc,
		mm:        mm,
		conf:      conf,
		store:     store,
		sessionID: sessionID,
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
	clone.ExpandedContext.BotAccessToken = app.BotAccessToken
	if expand == nil {
		return &clone, nil
	}

	// TODO: use the appropriate user's Mattermost OAuth2 token once
	// re-implemented, for now pass in the session token to make things work.
	if expand.AdminAccessToken != "" || expand.ActingUserAccessToken != "" {
		// Get the MM session
		if e.sessionID == "" {
			return nil, utils.NewUnauthorizedError("a user session is required")
		}
		if e.session == nil {
			session, err := utils.LoadSession(e.mm, e.sessionID, e.Context.ActingUserID)
			if err != nil {
				return nil, utils.NewUnauthorizedError(err)
			}
			e.session = session
		}

		if expand.AdminAccessToken != "" {
			if !app.GrantedPermissions.Contains(apps.PermissionActAsAdmin) {
				return nil, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsAdmin)
			}
			clone.ExpandedContext.AdminAccessToken = e.session.Token
		}
		if expand.ActingUserAccessToken != "" {
			if !app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
				return nil, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsUser)
			}
			clone.ExpandedContext.ActingUserAccessToken = e.session.Token
		}
	}

	clone.ExpandedContext.App = stripApp(app, expand.App)

	if expand.ActingUser != "" && e.ActingUserID != "" && e.ActingUser == nil {
		actingUser, err := e.mm.User.Get(e.ActingUserID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand acting user %s", e.ActingUserID)
		}
		e.ActingUser = actingUser
	}
	clone.ExpandedContext.ActingUser = stripUser(e.ActingUser, expand.ActingUser)

	if expand.Channel != "" && e.ChannelID != "" && e.Channel == nil {
		ch, err := e.mm.Channel.Get(e.ChannelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand channel %s", e.ChannelID)
		}
		e.Channel = ch
	}
	clone.ExpandedContext.Channel = stripChannel(e.Channel, expand.Channel)

	if expand.Post != "" && e.PostID != "" && e.Post == nil {
		post, err := e.mm.Post.GetPost(e.PostID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand post %s", e.PostID)
		}
		e.Post = post
	}
	clone.ExpandedContext.Post = stripPost(e.Post, expand.Post)

	if expand.RootPost != "" && e.RootPostID != "" && e.RootPost == nil {
		post, err := e.mm.Post.GetPost(e.RootPostID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand root post %s", e.RootPostID)
		}
		e.RootPost = post
	}
	clone.ExpandedContext.RootPost = stripPost(e.RootPost, expand.RootPost)

	if expand.Team != "" && e.TeamID != "" && e.Team == nil {
		team, err := e.mm.Team.Get(e.TeamID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand team %s", e.TeamID)
		}
		e.Team = team
	}
	clone.ExpandedContext.Team = stripTeam(e.Team, expand.Team)

	// TODO: expand Mentions, maybe replacing User?
	// https://mattermost.atlassian.net/browse/MM-30403

	if expand.User != "" && e.UserID != "" && e.User == nil {
		user, err := e.mm.User.Get(e.UserID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand user %s", e.UserID)
		}
		e.User = user
	}
	clone.ExpandedContext.User = stripUser(e.User, expand.User)

	if app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		conf := e.conf.GetConfig()
		if expand.OAuth2App != "" {
			clone.ExpandedContext.OAuth2.ClientID = app.RemoteOAuth2.ClientID
			clone.ExpandedContext.OAuth2.ClientSecret = app.RemoteOAuth2.ClientSecret
			clone.ExpandedContext.OAuth2.ConnectURL = conf.AppPath(app.AppID) + config.PathRemoteOAuth2Connect
			clone.ExpandedContext.OAuth2.CompleteURL = conf.AppPath(app.AppID) + config.PathRemoteOAuth2Complete
		}

		if expand.OAuth2User != "" && e.OAuth2.User == nil && e.ActingUserID != "" {
			var v interface{}
			err := e.store.OAuth2.GetUser(app.AppID, e.ActingUserID, &v)
			if err != nil && errors.Cause(err) != utils.ErrNotFound {
				return nil, errors.Wrapf(err, "failed to expand OAuth user %s", e.UserID)
			}
			clone.ExpandedContext.OAuth2.User = v
		}
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

	clone := apps.App{
		Manifest: apps.Manifest{
			AppID:   app.AppID,
			Version: app.Version,
		},
		WebhookSecret: app.WebhookSecret,
		BotUserID:     app.BotUserID,
		BotUsername:   app.BotUsername,
	}

	switch level {
	case apps.ExpandAll, apps.ExpandSummary:
		return &clone
	}
	return nil
}
