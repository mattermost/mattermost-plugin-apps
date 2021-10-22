package proxy

import (
	"path"

	"github.com/pkg/errors"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func contextForApp(app apps.App, base apps.Context, conf config.Config) apps.Context {
	out := base
	out.ExpandedContext = apps.ExpandedContext{}
	out.MattermostSiteURL = conf.MattermostSiteURL
	out.AppID = app.AppID
	out.AppPath = path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID))
	out.BotUserID = app.BotUserID
	out.BotAccessToken = app.BotAccessToken
	return out
}

var emptyCC = apps.Context{}

func (p *Proxy) expandContext(in Incoming, app apps.App, base *apps.Context, expand *apps.Expand) (apps.Context, error) {
	if base == nil {
		base = &apps.Context{}
	}
	conf, mm, _ := p.conf.Basic()

	cc := contextForApp(app, *base, conf)
	if expand == nil {
		// nothing more to do
		return cc, nil
	}

	client, err := getExpandClient(app, conf, mm, in)
	if err != nil {
		return emptyCC, err
	}

	var userSession *model.Session
	if expand.AdminAccessToken != "" {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsAdmin) {
			return emptyCC, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsAdmin)
		}
		err := utils.EnsureSysAdmin(p.conf.MattermostAPI(), cc.ActingUserID)
		if err != nil {
			return emptyCC, utils.NewForbiddenError("user is not a sysadmin")
		}
		// See if we can derive the admin token from the "base" context
		cc.AdminAccessToken = in.AdminAccessToken
		if cc.AdminAccessToken == "" {
			cc.AdminAccessToken = in.ActingUserAccessToken
		}
		// Try to obtain it from the present session
		if cc.AdminAccessToken == "" && in.SessionID != "" {
			userSession, err = utils.LoadSession(p.conf.MattermostAPI(), in.SessionID, in.ActingUserID)
			if err != nil {
				return emptyCC, utils.NewForbiddenError("failed to load user session")
			}
			cc.AdminAccessToken = userSession.Token
		}
		if cc.AdminAccessToken == "" {
			return cc, errors.New("admin access token is not available")
		}
	}

	if expand.ActingUserAccessToken != "" {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
			return emptyCC, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsUser)
		}
		cc.ActingUserAccessToken = in.ActingUserAccessToken
		if cc.ActingUserAccessToken == "" {
			if userSession == nil {
				var err error
				userSession, err = utils.LoadSession(p.conf.MattermostAPI(), in.SessionID, in.ActingUserID)
				if err != nil {
					return emptyCC, utils.NewForbiddenError("failed to load user session")
				}
			}
			cc.ActingUserAccessToken = userSession.Token
		}
		if cc.ActingUserAccessToken == "" {
			return emptyCC, utils.NewForbiddenError("acting user token is not available")
		}
	}

	cc.App = stripApp(app, expand.App)

	if expand.ActingUser != "" && base.ActingUserID != "" && base.ActingUser == nil {
		actingUser, err := client.GetUser(base.ActingUserID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand acting user %s", base.ActingUserID)
		}
		base.ActingUser = actingUser
	}
	cc.ActingUser = stripUser(base.ActingUser, expand.ActingUser)

	if expand.Channel != "" && base.ChannelID != "" && base.Channel == nil {
		ch, err := client.GetChannel(base.ChannelID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand channel %s", base.ChannelID)
		}
		base.Channel = ch
	}
	cc.Channel = stripChannel(base.Channel, expand.Channel)

	if expand.Post != "" && base.PostID != "" && base.Post == nil {
		post, err := client.GetPost(base.PostID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand post %s", base.PostID)
		}
		base.Post = post
	}
	cc.Post = stripPost(base.Post, expand.Post)

	if expand.RootPost != "" && base.RootPostID != "" && base.RootPost == nil {
		post, err := client.GetPost(base.RootPostID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand root post %s", base.RootPostID)
		}
		base.RootPost = post
	}
	cc.RootPost = stripPost(base.RootPost, expand.RootPost)

	if expand.Team != "" && base.TeamID != "" && base.Team == nil {
		team, err := client.GetTeam(base.TeamID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand team %s", base.TeamID)
		}
		base.Team = team
	}
	cc.Team = stripTeam(base.Team, expand.Team)

	// TODO: expand Mentions, maybe replacing User?
	// https://mattermost.atlassian.net/browse/MM-30403

	if expand.User != "" && base.UserID != "" && base.User == nil {
		user, err := client.GetUser(base.UserID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand user %s", base.UserID)
		}
		base.User = user
	}
	cc.User = stripUser(base.User, expand.User)

	cc.OAuth2 = apps.OAuth2Context{}
	if app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		if expand.OAuth2App != "" {
			cc.OAuth2.OAuth2App = app.RemoteOAuth2
			cc.OAuth2.ConnectURL = conf.AppURL(app.AppID) + appspath.RemoteOAuth2Connect
			cc.OAuth2.CompleteURL = conf.AppURL(app.AppID) + appspath.RemoteOAuth2Complete
		}

		if expand.OAuth2User != "" && base.OAuth2.User == nil && base.ActingUserID != "" {
			var v interface{}
			err := p.store.OAuth2.GetUser(app.BotUserID, base.ActingUserID, &v)
			if err != nil && !errors.Is(err, utils.ErrNotFound) {
				return emptyCC, errors.Wrapf(err, "failed to expand OAuth user %s", base.UserID)
			}
			cc.ExpandedContext.OAuth2.User = v
		}
	}

	if expand.Locale != "" {
		if cc.ActingUser != nil {
			cc.Locale = utils.GetLocaleWithUser(p.conf.MattermostConfig().Config(), cc.ActingUser)
		} else {
			cc.Locale = utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), cc.ActingUserID)
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

func stripApp(app apps.App, level apps.ExpandLevel) *apps.App {
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

func getExpandClient(app apps.App, conf config.Config, mm *pluginapi.Client, in Incoming) (mmclient.Client, error) {
	switch {
	case app.GrantedPermissions.Contains(apps.PermissionActAsAdmin):
		// If the app has admin permission anyway, use the RPC client for performance reasons
		return mmclient.NewRPCClient(mm), nil

	case app.GrantedPermissions.Contains(apps.PermissionActAsUser) && in.ActingUserID != "":
		// The OAuth2 token should be used here once it's implemented
		err := in.ensureUserToken(mm)
		if err != nil {
			return nil, err
		}
		return mmclient.NewHTTPClient(conf, in.ActingUserAccessToken), nil

	case app.GrantedPermissions.Contains(apps.PermissionActAsBot):
		return mmclient.NewHTTPClient(conf, app.BotAccessToken), nil

	default:
		return nil, utils.NewUnauthorizedError("apps without any ActAs* permission can't expand")
	}
}
