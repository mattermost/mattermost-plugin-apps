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
	out.ExpandedContext.MattermostSiteURL = conf.MattermostSiteURL
	out.ExpandedContext.DeveloperMode = conf.DeveloperMode
	out.UserAgentContext.AppID = app.AppID
	out.ExpandedContext.AppPath = path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID))

	out.ExpandedContext.BotUserID = app.BotUserID

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) && app.BotUserID != "" {
		botAccessToken, err := p.getBotAccessToken(r, app)
		if err != nil {
			return emptyCC, err
		}
		out.ExpandedContext.BotAccessToken = botAccessToken
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

	cc.App, err = expandApp(app, expand.App)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand app")
	}

	cc.ActingUser, err = expandUser(client, expand.ActingUser, base.ActingUserID, apps.ExpandID)
	if err != nil {
		return emptyCC, errors.Wrapf(err, "failed to expand acting user")
	}

	if expand.ActingUserAccessToken == apps.ExpandAll {
		if !app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
			return emptyCC, utils.NewForbiddenError("%s does not have permission to %s", app.AppID, apps.PermissionActAsUser)
		}

		cc.ActingUserAccessToken, err = r.UserAccessToken()
		if err != nil {
			return emptyCC, errors.New("failed to load user session")
		}
	}

	if expand.Locale == apps.ExpandSummary || expand.Locale == apps.ExpandAll {
		if cc.ActingUser != nil {
			cc.Locale = utils.GetLocaleWithUser(p.conf.MattermostConfig().Config(), cc.ActingUser)
		} else {
			cc.Locale = utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), cc.ActingUserID)
		}
	}

	cc.Channel, err = expandChannel(client, expand.Channel, base.ChannelID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand channel")
	}

	cc.ChannelMember, err = expandChannelMember(client, expand.ChannelMember, base.ChannelID, base.ActingUserID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand channel membership")
	}

	cc.Team, err = expandTeam(client, expand.Team, base.TeamID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand team")
	}

	cc.TeamMember, err = expandTeamMember(client, expand.TeamMember, base.TeamID, base.ActingUserID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand team membership")
	}

	cc.Post, err = expandPost(client, expand.Post, base.PostID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand post")
	}

	cc.RootPost, err = expandPost(client, expand.RootPost, base.RootPostID)
	if err != nil {
		return emptyCC, errors.Wrap(err, "failed to expand root post ")
	}

	// TODO: expand Mentions, maybe replacing User?
	// https://mattermost.atlassian.net/browse/MM-30403
	cc.User, err = expandUser(client, expand.User, base.UserID, apps.ExpandNone)
	if err != nil {
		return emptyCC, errors.Wrapf(err, "failed to expand user")
	}

	cc.OAuth2 = apps.OAuth2Context{}
	if app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		if expand.OAuth2App == apps.ExpandSummary || expand.OAuth2App == apps.ExpandAll {
			cc.OAuth2.OAuth2App = app.RemoteOAuth2
			cc.OAuth2.ConnectURL = conf.AppURL(app.AppID) + appspath.RemoteOAuth2Connect
			cc.OAuth2.CompleteURL = conf.AppURL(app.AppID) + appspath.RemoteOAuth2Complete
		}

		if (expand.OAuth2User == apps.ExpandSummary || expand.OAuth2User == apps.ExpandAll) && base.OAuth2.User == nil && base.ActingUserID != "" {
			var data []byte
			data, err = p.appservices.GetOAuth2User(r, app.AppID, base.ActingUserID)
			if err != nil {
				return emptyCC, errors.Wrap(err, "failed to expand OAuth user")
			}
			if len(data) > 0 {
				var v interface{}
				if err = json.Unmarshal(data, &v); err != nil {
					return emptyCC, errors.Wrapf(err, "failed unmarshal OAuth2 User %s", base.UserID)
				}
				cc.ExpandedContext.OAuth2.User = v
			}
		}
	}

	if app.DeployType != apps.DeployBuiltin {
		// Cleanup fields for app
		cc.UserAgentContext.ChannelID = ""
		cc.UserAgentContext.TeamID = ""
		cc.UserAgentContext.RootPostID = ""
		cc.UserAgentContext.PostID = ""
		cc.UserID = ""
		cc.ActingUserID = ""
	}

	return cc, nil
}

func expandUser(client mmclient.Client, level apps.ExpandLevel, userID string, defaultLevel apps.ExpandLevel) (*model.User, error) {
	if userID == "" {
		return nil, nil
	}

	if level == apps.ExpandDefault {
		level = defaultLevel
	}

	switch level {
	case apps.ExpandNone:
		return nil, nil
	case apps.ExpandID:
		return &model.User{
			Id: userID,
		}, nil
	case apps.ExpandSummary:
		user, err := client.GetUser(userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand user %s", userID)
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
		}, nil
	case apps.ExpandAll:
		user, err := client.GetUser(userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to expand user %s", userID)
		}

		user.Sanitize(map[string]bool{
			"passwordupdate": true,
			"authservice":    true,
		})
		return user, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandChannel(client mmclient.Client, level apps.ExpandLevel, channelID string) (*model.Channel, error) {
	if channelID == "" {
		return nil, nil
	}

	switch level {
	case apps.ExpandNone,
		apps.ExpandDefault:
		return nil, nil
	case apps.ExpandID:
		return &model.Channel{
			Id: channelID,
		}, nil
	case apps.ExpandSummary:
		channel, err := client.GetChannel(channelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel %s", channelID)
		}

		return &model.Channel{
			Id:          channel.Id,
			DeleteAt:    channel.DeleteAt,
			TeamId:      channel.TeamId,
			Type:        channel.Type,
			DisplayName: channel.DisplayName,
			Name:        channel.Name,
		}, nil
	case apps.ExpandAll:
		channel, err := client.GetChannel(channelID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get channel %s", channelID)
		}

		return channel, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandChannelMember(client mmclient.Client, level apps.ExpandLevel, channelID, userID string) (*model.ChannelMember, error) {
	if channelID == "" || userID == "" {
		return nil, nil
	}

	switch level {
	case apps.ExpandNone, apps.ExpandDefault, apps.ExpandID:
		return nil, nil
	case apps.ExpandSummary, apps.ExpandAll:
		cm, err := client.GetChannelMember(channelID, userID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get channel membership")
		}

		return cm, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandTeam(client mmclient.Client, level apps.ExpandLevel, teamID string) (*model.Team, error) {
	if teamID == "" {
		return nil, nil
	}

	switch level {
	case apps.ExpandNone,
		apps.ExpandDefault:
		return nil, nil
	case apps.ExpandID:
		return &model.Team{
			Id: teamID,
		}, nil
	case apps.ExpandSummary:
		team, err := client.GetTeam(teamID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get team %s", teamID)
		}

		return &model.Team{
			Id:          team.Id,
			DisplayName: team.DisplayName,
			Name:        team.Name,
			Description: team.Description,
			Email:       team.Email,
			Type:        team.Type,
		}, nil
	case apps.ExpandAll:
		team, err := client.GetTeam(teamID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get team %s", teamID)
		}

		sanitized := *team
		sanitized.Sanitize()
		return &sanitized, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandTeamMember(client mmclient.Client, level apps.ExpandLevel, teamID, userID string) (*model.TeamMember, error) {
	if teamID == "" || userID == "" {
		return nil, nil
	}

	switch level {
	case apps.ExpandNone, apps.ExpandDefault, apps.ExpandID:
		return nil, nil
	case apps.ExpandSummary, apps.ExpandAll:
		cm, err := client.GetTeamMember(teamID, userID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get team membership %s", teamID)
		}

		return cm, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandPost(client mmclient.Client, level apps.ExpandLevel, postID string) (*model.Post, error) {
	if postID == "" {
		return nil, nil
	}

	switch level {
	case apps.ExpandNone,
		apps.ExpandDefault:
		return nil, nil
	case apps.ExpandID:
		return &model.Post{
			Id: postID,
		}, nil
	case apps.ExpandSummary:
		post, err := client.GetPost(postID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get post %s", postID)
		}

		return &model.Post{
			Id:        post.Id,
			Type:      post.Type,
			UserId:    post.UserId,
			ChannelId: post.ChannelId,
			RootId:    post.RootId,
			Message:   post.Message,
		}, nil
	case apps.ExpandAll:
		post, err := client.GetPost(postID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get post %s", postID)
		}

		post.SanitizeProps()
		return post, nil

	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
}

func expandApp(app apps.App, level apps.ExpandLevel) (*apps.App, error) {
	switch level {
	case apps.ExpandNone, apps.ExpandDefault, apps.ExpandID:
		return nil, nil
	case apps.ExpandAll, apps.ExpandSummary:
		return &apps.App{
			Manifest: apps.Manifest{
				AppID:   app.AppID,
				Version: app.Version,
			},
			WebhookSecret: app.WebhookSecret,
			BotUserID:     app.BotUserID,
			BotUsername:   app.BotUsername,
		}, nil
	default:
		return nil, errors.Errorf("unknown expand type %q", level)
	}
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
