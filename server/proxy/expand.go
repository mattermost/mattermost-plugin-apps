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

type expandFunc func(apps.ExpandLevel) error

type expander struct {
	apps.Context
	app    *apps.App
	client mmclient.Client
	r      *incoming.Request
	proxy  *Proxy
}

// expandContext performs `expand` on a `context` of an outgoing CallRequest. It
// ensures a fresh/correct ExpandedContext. It preserves the UserAgentContext
// since it is used during expand, which cleans it out before returning. It
// relies on `r` configured with correct source and destination app IDs, as well
// as the acting user data. (destination) `app` is passed in as a shortcut since
// it's already available in all callers.
func (p *Proxy) expandContext(r *incoming.Request, app *apps.App, cc *apps.Context, expand *apps.Expand) (*apps.Context, error) {
	if cc == nil {
		cc = &apps.Context{}
	}

	client, err := p.getExpandClient(r)
	if err != nil {
		return nil, err
	}
	e := &expander{
		app:     app,
		Context: *cc,
		proxy:   p,
		client:  client,
		r:       r,
	}

	conf := r.Config().Get()
	e.ExpandedContext = apps.ExpandedContext{
		AppPath:           path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID)),
		BotUserID:         app.BotUserID,
		DeveloperMode:     conf.DeveloperMode,
		MattermostSiteURL: conf.MattermostSiteURL,
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) && app.BotUserID != "" {
		botAccessToken, err := p.getBotAccessToken(r, app)
		if err != nil {
			return nil, err
		}
		e.ExpandedContext.BotAccessToken = botAccessToken
	}

	return e.expand(expand)
}

// expand expands the context according to the expand parameter, and the IDs
// passed in, mostly in the UserAgentContext part of the request. It returns a
// "clean" context, ready to be passed down to the app.
func (e *expander) expand(expand *apps.Expand) (*apps.Context, error) {
	if expand == nil {
		expand = &apps.Expand{}
	}

	// TODO: expand Mentions, maybe replacing User?
	// https://mattermost.atlassian.net/browse/MM-30403
	for _, step := range []struct {
		name           string
		f              expandFunc
		requestedLevel apps.ExpandLevel
		expandableAs   []apps.ExpandLevel
		defaultExpand  apps.ExpandLevel
		err            error
	}{
		{
			name:           "acting_user_access_token",
			requestedLevel: expand.ActingUserAccessToken,
			f:              e.expandActingUserAccessToken,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Required(),
			err:            utils.NewInvalidError(`invalid expand: "acting_user_access_token" must be "all" or empty`),
		}, {
			name:           "acting_user",
			requestedLevel: expand.ActingUser,
			f:              e.expandUser(&e.ExpandedContext.ActingUser, e.r.ActingUserID()),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Required(),
		}, {
			name:           "app",
			requestedLevel: expand.App,
			f:              e.expandApp,
			expandableAs:   []apps.ExpandLevel{apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Required(),
		}, {
			name:           "channel_member",
			requestedLevel: expand.ChannelMember,
			f:              e.expandChannelMember,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "channel",
			requestedLevel: expand.Channel,
			f:              e.expandChannel,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			// Locale must be expanded after acting_user
			name:           "locale",
			requestedLevel: expand.Locale,
			f:              e.expandLocale,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "oauth2_app",
			requestedLevel: expand.OAuth2App,
			f:              e.expandOAuth2App,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Required(),
			err:            utils.NewInvalidError(`invalid expand: "oauth2_app" must be "all" or empty`),
		}, {
			name:           "oauth2_user",
			requestedLevel: expand.OAuth2User,
			f:              e.expandOAuth2User,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Required(),
			err:            utils.NewInvalidError(`invalid expand: "oauth2_user" must be "all" or empty`),
		}, {
			name:           "post",
			requestedLevel: expand.Post,
			f:              e.expandPost(&e.ExpandedContext.Post, e.UserAgentContext.PostID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "root_post",
			requestedLevel: expand.RootPost,
			f:              e.expandPost(&e.ExpandedContext.RootPost, e.UserAgentContext.RootPostID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "team_member",
			requestedLevel: expand.TeamMember,
			f:              e.expandTeamMember,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "team",
			requestedLevel: expand.Team,
			f:              e.expandTeam,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		}, {
			name:           "user",
			requestedLevel: expand.User,
			f:              e.expandUser(&e.ExpandedContext.User, e.UserID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
			defaultExpand:  apps.ExpandNone.Optional(),
		},
	} {
		level, err := apps.ParseExpandLevel(string(step.requestedLevel), step.defaultExpand)
		if err != nil {
			return nil, err
		}

		// Check the requested expand level to see if it's valid and if there is
		// anything to do.
		l := level.Level()
		doExpand := false
		for _, expandable := range step.expandableAs {
			if l == expandable {
				doExpand = true
				break
			}
		}
		if !doExpand {
			continue
		}

		// Execute the expand function, skip the error unless the field is
		// required.
		if err := step.f(level.Level()); err != nil && level.IsRequired() {
			return nil, errors.Wrap(err, "failed to expand required "+step.name)
		}
	}

	// Cross-check the results for consistency.
	if err := e.consistencyCheck(); err != nil {
		return nil, errors.Wrap(err, "failed post-expand consistency check")
	}

	// Cleanup fields that must not go to the app.
	e.Context.UserAgentContext = apps.UserAgentContext{}
	e.Context.UserID = ""
	return &e.Context, nil
}

func (e *expander) expandActingUserAccessToken(level apps.ExpandLevel) error {
	to := e.app
	if !to.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		return utils.NewForbiddenError("%s does not have permission to %s", to.AppID, apps.PermissionActAsUser)
	}

	token, err := e.r.ActingUserAccessTokenForDestination()
	if err != nil {
		return err
	}
	e.ExpandedContext.ActingUserAccessToken = token
	return nil
}

func (e *expander) expandUser(userPtr **model.User, userID string) expandFunc {
	return func(level apps.ExpandLevel) error {
		if userID == "" {
			return errors.New("no user ID to expand")
		}
		if userPtr == nil {
			return errors.New("internal unreachable error: nil userPtr")
		}

		switch level {
		case apps.ExpandID:
			// The GetUser API invocation here serves as the access check, but
			// it is not needed for the acting user, so skip it, as an
			// optimization.
			if userID != e.r.ActingUserID() {
				if _, err := e.client.GetUser(userID); err != nil {
					return errors.Wrapf(err, "id: %s", userID)
				}
			}
			*userPtr = &model.User{
				Id: userID,
			}

		case apps.ExpandSummary:
			u, err := e.client.GetUser(userID)
			if err != nil {
				return errors.Wrapf(err, "id: %s", userID)
			}
			*userPtr = &model.User{
				BotDescription: u.BotDescription,
				DeleteAt:       u.DeleteAt,
				Email:          u.Email,
				FirstName:      u.FirstName,
				Id:             u.Id,
				IsBot:          u.IsBot,
				LastName:       u.LastName,
				Locale:         u.Locale,
				Nickname:       u.Nickname,
				Roles:          u.Roles,
				Timezone:       u.Timezone,
				Username:       u.Username,
			}

		case apps.ExpandAll:
			u, err := e.client.GetUser(userID)
			if err != nil {
				return errors.Wrapf(err, "id: %s", userID)
			}
			u.Sanitize(map[string]bool{
				"email":    true,
				"fullname": true,
			})
			u.AuthData = nil
			*userPtr = u
		}
		return nil
	}
}

func (e *expander) expandApp(level apps.ExpandLevel) error {
	switch level {
	case apps.ExpandSummary:
		e.ExpandedContext.App = &apps.App{
			Manifest: apps.Manifest{
				AppID:   e.app.AppID,
				Version: e.app.Version,
			},
			BotUserID:   e.app.BotUserID,
			BotUsername: e.app.BotUsername,
		}

	case apps.ExpandAll:
		e.ExpandedContext.App = &apps.App{
			Manifest: apps.Manifest{
				AppID:   e.app.AppID,
				Version: e.app.Version,
			},
			BotUserID:     e.app.BotUserID,
			BotUsername:   e.app.BotUsername,
			DeployType:    e.app.DeployType,
			WebhookSecret: e.app.WebhookSecret,
		}
	}
	return nil
}

func (e *expander) expandChannelMember(level apps.ExpandLevel) error {
	channelID := e.UserAgentContext.ChannelID
	userID := e.UserID
	if userID == "" {
		userID = e.r.ActingUserID()
	}
	if userID == "" || channelID == "" {
		return errors.New("no user ID or channel ID to expand")
	}

	cm, err := e.client.GetChannelMember(channelID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get channel membership")
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.ChannelMember = &model.ChannelMember{
			UserId:    cm.UserId,
			ChannelId: cm.ChannelId,
		}

	case apps.ExpandSummary,
		apps.ExpandAll:
		e.ExpandedContext.ChannelMember = cm
	}
	return nil
}

func (e *expander) expandChannel(level apps.ExpandLevel) error {
	channelID := e.UserAgentContext.ChannelID
	if channelID == "" {
		return errors.New("no channel ID to expand")
	}
	channel, err := e.client.GetChannel(channelID)
	if err != nil {
		return errors.Wrap(err, "id: "+channelID)
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.Channel = &model.Channel{
			Id: channel.Id,
		}

	case apps.ExpandSummary:
		e.ExpandedContext.Channel = &model.Channel{
			Id:          channel.Id,
			DeleteAt:    channel.DeleteAt,
			TeamId:      channel.TeamId,
			Type:        channel.Type,
			DisplayName: channel.DisplayName,
			Name:        channel.Name,
		}

	case apps.ExpandAll:
		e.ExpandedContext.Channel = channel
	}
	return nil
}

func (e *expander) expandTeam(level apps.ExpandLevel) error {
	teamID := e.UserAgentContext.TeamID
	if teamID == "" {
		return errors.New("no team ID to expand")
	}
	team, err := e.client.GetTeam(teamID)
	if err != nil {
		return errors.Wrapf(err, "failed to get team %s", teamID)
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.Team = &model.Team{
			Id: team.Id,
		}
	case apps.ExpandSummary:
		e.ExpandedContext.Team = &model.Team{
			Id:          team.Id,
			DisplayName: team.DisplayName,
			Name:        team.Name,
			Description: team.Description,
			Email:       team.Email,
			Type:        team.Type,
		}

	case apps.ExpandAll:
		sanitized := *team
		sanitized.Sanitize()
		e.ExpandedContext.Team = &sanitized
	}

	return nil
}

func (e *expander) expandTeamMember(level apps.ExpandLevel) error {
	teamID := e.UserAgentContext.TeamID
	userID := e.UserID
	if userID == "" {
		userID = e.r.ActingUserID()
	}
	if userID == "" || teamID == "" {
		return errors.New("no user ID or channel ID to expand")
	}
	tm, err := e.client.GetTeamMember(teamID, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get team membership")
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.TeamMember = &model.TeamMember{
			UserId: tm.UserId,
			TeamId: tm.TeamId,
		}
	case apps.ExpandSummary, apps.ExpandAll:
		e.ExpandedContext.TeamMember = tm
	}
	return nil
}

func (e *expander) expandPost(postPtr **model.Post, postID string) expandFunc {
	return func(level apps.ExpandLevel) error {
		if postID == "" {
			return errors.New("no post ID to expand")
		}
		post, err := e.client.GetPost(postID)
		if err != nil {
			return errors.Wrapf(err, "failed to get post %s", postID)
		}

		switch level {
		case apps.ExpandID:
			*postPtr = &model.Post{
				Id: post.Id,
			}

		case apps.ExpandSummary:
			*postPtr = &model.Post{
				Id:        post.Id,
				Type:      post.Type,
				UserId:    post.UserId,
				ChannelId: post.ChannelId,
				RootId:    post.RootId,
				Message:   post.Message,
			}

		case apps.ExpandAll:
			post.SanitizeProps()
			*postPtr = post
		}
		return nil
	}
}

func (e *expander) expandLocale(level apps.ExpandLevel) error {
	if e.ExpandedContext.ActingUser != nil {
		e.ExpandedContext.Locale = utils.GetLocaleWithUser(e.r.Config().MattermostConfig().Config(), e.ExpandedContext.ActingUser)
	} else {
		e.ExpandedContext.Locale = utils.GetLocale(e.r.Config().MattermostAPI(), e.r.Config().MattermostConfig().Config(), e.r.ActingUserID())
	}
	return nil
}

func (e *expander) expandOAuth2App(level apps.ExpandLevel) error {
	to, err := e.proxy.GetInstalledApp(e.r.Destination(), true)
	if err != nil {
		return err
	}
	if !to.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewForbiddenError("%s does not have permission to %s", to.AppID, apps.PermissionRemoteOAuth2)
	}

	conf := e.r.Config().Get()
	e.ExpandedContext.OAuth2.OAuth2App = to.RemoteOAuth2
	e.ExpandedContext.OAuth2.ConnectURL = conf.AppURL(to.AppID) + appspath.RemoteOAuth2Connect
	e.ExpandedContext.OAuth2.CompleteURL = conf.AppURL(to.AppID) + appspath.RemoteOAuth2Complete
	return nil
}

func (e *expander) expandOAuth2User(level apps.ExpandLevel) error {
	userID := e.r.ActingUserID()
	if userID == "" {
		return errors.New("no acting user id to expand")
	}
	to, err := e.proxy.GetInstalledApp(e.r.Destination(), true)
	if err != nil {
		return err
	}
	if !to.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewForbiddenError("%s does not have permission to %s", to.AppID, apps.PermissionRemoteOAuth2)
	}

	appServicesRequest := e.r.WithSourceAppID(e.r.Destination())
	data, err := e.proxy.appservices.GetOAuth2User(appServicesRequest)
	if err != nil {
		return errors.Wrap(err, "user_id: "+userID)
	}
	if len(data) == 0 || string(data) == "{}" {
		return errors.Wrap(err, "no data for user_id: "+userID)
	}

	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return errors.Wrapf(err, "user_id: "+userID)
	}
	e.ExpandedContext.OAuth2.User = v
	return nil
}

func (p *Proxy) getExpandClient(r *incoming.Request) (mmclient.Client, error) {
	if p.expandClientOverride != nil {
		return p.expandClientOverride, nil
	}
	app, err := p.getEnabledDestination(r)
	if err != nil {
		return nil, err
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

func (p *Proxy) getBotAccessToken(r *incoming.Request, app *apps.App) (string, error) {
	session, err := p.sessionService.GetOrCreate(r, app.AppID, app.BotUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get bot session")
	}

	return session.Token, nil
}

func (e *expander) consistencyCheck() error {
	if e.ExpandedContext.Post != nil {
		if e.ExpandedContext.Post.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded post's channel ID %s is different from user agent context %s",
				e.ExpandedContext.Post.ChannelId, e.UserAgentContext.ChannelID)
		}
	}

	if e.ExpandedContext.Channel != nil {
		if e.ExpandedContext.Channel.Type != model.ChannelTypeDirect && e.ExpandedContext.Channel.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded channel's team ID %s is different from user agent context %s",
				e.ExpandedContext.Channel.TeamId, e.UserAgentContext.TeamID)
		}
	}

	if e.ExpandedContext.ChannelMember != nil {
		if e.ExpandedContext.ChannelMember.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded channel member's channel ID %s is different from user agent context %s",
				e.ExpandedContext.ChannelMember.ChannelId, e.UserAgentContext.ChannelID)
		}
		if e.ExpandedContext.ChannelMember.UserId != e.r.ActingUserID() && e.ExpandedContext.ChannelMember.UserId != e.UserID {
			return errors.Errorf("expanded channel member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	if e.ExpandedContext.TeamMember != nil {
		if e.ExpandedContext.TeamMember.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded team member's team ID %s is different from user agent context %s",
				e.ExpandedContext.TeamMember.TeamId, e.UserAgentContext.TeamID)
		}
		if e.ExpandedContext.TeamMember.UserId != e.r.ActingUserID() && e.ExpandedContext.TeamMember.UserId != e.UserID {
			return errors.Errorf("expanded team member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	return nil
}
