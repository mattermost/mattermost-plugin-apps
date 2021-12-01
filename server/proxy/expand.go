package proxy

import (
	"encoding/json"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) contextForApp(r *incoming.Request, app apps.App, base apps.Context) (apps.Context, error) {
	conf := p.conf.Get()

	out := base
	out.ExpandedContext = apps.ExpandedContext{}
	out.MattermostSiteURL = conf.MattermostSiteURL
	out.DeveloperMode = conf.DeveloperMode
	out.AppID = app.AppID
	out.AppPath = path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID))

	out.BotUserID = app.BotUserID

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) && app.BotUserID != "" {
		botAccessToken, err := p.getBotAccessToken(r, app)
		if err != nil {
			return emptyCC, err
		}
		out.BotAccessToken = botAccessToken
	}

	return out, nil
}

var emptyCC = apps.Context{}

func (p *Proxy) expandContext(r *incoming.Request, app apps.App, base *apps.Context, expand *apps.Expand) (apps.Context, error) {
	if base == nil {
		base = &apps.Context{}
	}
	conf := p.conf.Get()

	cc, err := p.contextForApp(r, app, *base)
	if err != nil {
		return emptyCC, err
	}

	if expand == nil {
		// nothing more to do
		return cc, nil
	}

	client, err := p.getExpandClient(r, app)
	if err != nil {
		return emptyCC, err
	}

	if expand.ActingUserAccessToken != "" {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
			return emptyCC, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsUser)
		}

		cc.ActingUserAccessToken, err = r.UserAccessToken()
		if err != nil {
			return emptyCC, errors.New("failed to load user session")
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

	if expand.ChannelMember != "" && base.ChannelID != "" && base.ActingUserID != "" && base.ChannelMember == nil {
		cm, err := client.GetChannelMember(base.ChannelID, base.ActingUserID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand channel membership %s", base.ChannelID)
		}
		base.ChannelMember = cm
	}
	cc.ChannelMember = stripChannelMember(base.ChannelMember, expand.ChannelMember)

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

	if expand.TeamMember != "" && base.TeamID != "" && base.ActingUserID != "" && base.TeamMember == nil {
		tm, err := client.GetTeamMember(base.TeamID, base.ActingUserID)
		if err != nil {
			return emptyCC, errors.Wrapf(err, "failed to expand team membership %s", base.TeamID)
		}
		base.TeamMember = tm
	}
	cc.TeamMember = stripTeamMember(base.TeamMember, expand.TeamMember)

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
			data, err := p.store.OAuth2.GetUser(r, app.AppID, base.ActingUserID)
			if err != nil && !errors.Is(err, utils.ErrNotFound) {
				return emptyCC, errors.Wrapf(err, "failed to expand OAuth user %s", base.UserID)
			}

			var v interface{}
			if err = json.Unmarshal(data, v); err != nil {
				return emptyCC, errors.Wrapf(err, "failed unmarshal OAuth2 User %s", base.UserID)
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

func stripChannelMember(cm *model.ChannelMember, level apps.ExpandLevel) *model.ChannelMember {
	if cm == nil || (level != apps.ExpandAll && level != apps.ExpandSummary) {
		return nil
	}
	return cm
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

func stripTeamMember(tm *model.TeamMember, level apps.ExpandLevel) *model.TeamMember {
	if tm == nil || (level != apps.ExpandAll && level != apps.ExpandSummary) {
		return nil
	}
	return tm
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

func (p *Proxy) getExpandClient(r *incoming.Request, app apps.App) (mmclient.Client, error) {
	if p.expandClientOverride != nil {
		return p.expandClientOverride, nil
	}

	conf := p.conf.Get()

	switch {
	case app.DeployType == apps.DeployBuiltin:
		return mmclient.NewRPCClient(p.conf.MattermostAPI()), nil

	case app.GrantedPermissions.Contains(apps.PermissionActAsUser) && r.ActingUserID() != "":
		return r.GetMMClient()

	case app.GrantedPermissions.Contains(apps.PermissionActAsBot):
		accessToken, err := p.getBotAccessToken(r, app)
		if err != nil {
			return nil, err
		}
		return mmclient.NewHTTPClient(conf, accessToken), nil

	default:
		return nil, utils.NewUnauthorizedError("apps without any ActAs* permission can't expand")
	}
}

func (p *Proxy) getBotAccessToken(r *incoming.Request, app apps.App) (string, error) {
	session, err := p.sessionService.GetOrCreate(r, app.AppID, app.BotUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get bot session")
	}

	return session.Token, nil
}
