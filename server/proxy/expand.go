package proxy

import (
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func forApp(app *apps.App, cc apps.Context, conf config.Config) apps.Context {
	cc.MattermostSiteURL = conf.MattermostSiteURL
	cc.AppID = app.AppID
	cc.AppPath = path.Join(conf.PluginURLPath, mmclient.PathApps, string(app.AppID))
	cc.BotUserID = app.BotUserID
	cc.BotAccessToken = app.BotAccessToken
	return cc
}

func (p *Proxy) expandContext(base *apps.Context, app *apps.App, expand *apps.Expand) (apps.Context, error) {
	if base == nil {
		base = &apps.Context{}
	}
	conf := p.conf.GetConfig()
	cc := forApp(app, *base, conf)

	if expand == nil {
		// nothing more to do
		cc.ExpandedContext = apps.ExpandedContext{}
		return cc, nil
	}

	cc.AdminAccessToken = ""
	if expand.AdminAccessToken != "" {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsAdmin) {
			return apps.Context{}, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsAdmin)
		}
		if base.AdminAccessToken == "" {
			return apps.Context{}, utils.NewForbiddenError("admin token is not available")
		}
		cc.AdminAccessToken = base.AdminAccessToken
	}

	cc.ActingUserAccessToken = ""
	if expand.ActingUserAccessToken != "" {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
			return apps.Context{}, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsUser)
		}
		if base.ActingUserAccessToken == "" {
			return apps.Context{}, utils.NewForbiddenError("acting user token is not available")
		}
		cc.ActingUserAccessToken = base.ActingUserAccessToken
	}

	cc.App = stripApp(app, expand.App)

	if expand.ActingUser != "" && base.ActingUserID != "" && base.ActingUser == nil {
		actingUser, err := p.mm.User.Get(base.ActingUserID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand acting user %s", base.ActingUserID)
		}
		base.ActingUser = actingUser
	}
	cc.ActingUser = stripUser(base.ActingUser, expand.ActingUser)

	if expand.Channel != "" && base.ChannelID != "" && base.Channel == nil {
		ch, err := p.mm.Channel.Get(base.ChannelID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand channel %s", base.ChannelID)
		}
		base.Channel = ch
	}
	cc.Channel = stripChannel(base.Channel, expand.Channel)

	if expand.Post != "" && base.PostID != "" && base.Post == nil {
		post, err := p.mm.Post.GetPost(base.PostID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand post %s", base.PostID)
		}
		base.Post = post
	}
	cc.Post = stripPost(base.Post, expand.Post)

	if expand.RootPost != "" && base.RootPostID != "" && base.RootPost == nil {
		post, err := p.mm.Post.GetPost(base.RootPostID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand root post %s", base.RootPostID)
		}
		base.RootPost = post
	}
	cc.RootPost = stripPost(base.RootPost, expand.RootPost)

	if expand.Team != "" && base.TeamID != "" && base.Team == nil {
		team, err := p.mm.Team.Get(base.TeamID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand team %s", base.TeamID)
		}
		base.Team = team
	}
	cc.Team = stripTeam(base.Team, expand.Team)

	if expand.User != "" && base.UserID != "" && base.User == nil {
		user, err := p.mm.User.Get(base.UserID)
		if err != nil {
			return apps.Context{}, errors.Wrapf(err, "failed to expand user %s", base.UserID)
		}
		base.User = user
	}
	cc.User = stripUser(base.User, expand.User)

	cc.OAuth2 = apps.OAuth2Context{}
	if app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		if expand.OAuth2App != "" {
			cc.OAuth2.ClientID = app.RemoteOAuth2.ClientID
			cc.OAuth2.ClientSecret = app.RemoteOAuth2.ClientSecret
			cc.OAuth2.ConnectURL = conf.AppURL(app.AppID) + config.PathRemoteOAuth2Connect
			cc.OAuth2.CompleteURL = conf.AppURL(app.AppID) + config.PathRemoteOAuth2Complete
		}

		if expand.OAuth2User != "" && base.OAuth2.User == nil && base.ActingUserID != "" {
			var v interface{}
			err := p.store.OAuth2.GetUser(app.BotUserID, base.ActingUserID, &v)
			if err != nil && !errors.Is(err, utils.ErrNotFound) {
				return apps.Context{}, errors.Wrapf(err, "failed to expand OAuth user %s", base.UserID)
			}
			cc.ExpandedContext.OAuth2.User = v
		}
	}

	return cc, nil
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
