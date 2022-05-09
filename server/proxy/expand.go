package proxy

import (
	"encoding/json"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

type expandFunc func(apps.ExpandLevel) error

type expander struct {
	apps.Context
	app         *apps.App
	client      mmclient.Client
	r           *incoming.Request
	appservices appservices.Service
}

func (p *Proxy) expandContext(r *incoming.Request, app apps.App, cc *apps.Context, expand *apps.Expand) (*apps.Context, error) {
	if cc == nil {
		cc = &apps.Context{}
	}

	e, err := p.newExpander(r, app, cc, p.appservices)
	if err != nil {
		return nil, err
	}

	return e.expand(expand)
}

// newExpander creates a context expander. It ensures a fresh/correct
// ExpandedContext, ActingUserID, and UserAgentContext.AppID. It preserves the
// UserAgentContext since it is used during expand, which cleans it out before
// returning.
func (p *Proxy) newExpander(r *incoming.Request, app apps.App, in *apps.Context, appservices appservices.Service) (*expander, error) {
	client, err := p.getExpandClient(r, app)
	if err != nil {
		return nil, err
	}

	e := &expander{
		Context:     *in,
		app:         &app,
		appservices: appservices,
		client:      client,
		r:           r,
	}

	conf := r.Config().Get()
	e.ExpandedContext = apps.ExpandedContext{
		AppPath:           path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID)),
		BotUserID:         app.BotUserID,
		DeveloperMode:     conf.DeveloperMode,
		MattermostSiteURL: conf.MattermostSiteURL,
	}

	// For a call the AppID comes in the CallRequest now, so copy it to
	// UserAgentContext.
	e.UserAgentContext.AppID = app.AppID

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) && app.BotUserID != "" {
		botAccessToken, err := p.getBotAccessToken(r, app)
		if err != nil {
			return nil, err
		}
		e.ExpandedContext.BotAccessToken = botAccessToken
	}

	return e, nil
}

// expand expands the context according to the expand parameter, and the IDs
// passed in, mostly in the UserAgentContext part of the request. It returns a
// "clean" context, ready to be passed down to the app.
func (e *expander) expand(expand *apps.Expand) (*apps.Context, error) {
	knownLevels := []apps.ExpandLevel{apps.ExpandDefault, apps.ExpandNone, apps.ExpandID, apps.ExpandSummary, apps.ExpandAll}

	if expand == nil {
		expand = &apps.Expand{}
	}

	// TODO: expand Mentions, maybe replacing User?
	// https://mattermost.atlassian.net/browse/MM-30403
	for _, step := range []struct {
		name              string
		f                 expandFunc
		requestedLevel    apps.ExpandLevel
		expandableAs      []apps.ExpandLevel
		optionalByDefault bool
		err               error
	}{
		{
			name:           "acting_user_access_token",
			requestedLevel: expand.ActingUserAccessToken,
			f:              e.expandActingUserAccessToken,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			err:            utils.NewInvalidError(`invalid expand: "acting_user_access_token" must be "all" or empty`),
		}, {
			name:           "acting_user",
			requestedLevel: expand.ActingUser,
			f:              e.expandUser(&e.ExpandedContext.ActingUser, e.ActingUserID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "app",
			requestedLevel: expand.App,
			f:              e.expandApp,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "channel_member",
			requestedLevel: expand.ChannelMember,
			f:              e.expandChannelMember,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "channel",
			requestedLevel: expand.Channel,
			f:              e.expandChannel,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			// Locale must be expanded after acting_user
			name:           "locale",
			requestedLevel: expand.Locale,
			f:              e.expandLocale,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "oauth2_app",
			requestedLevel: expand.OAuth2App,
			f:              e.expandOAuth2App,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			err:            utils.NewInvalidError(`invalid expand: "oauth2_app" must be "all" or empty`),
		}, {
			name:           "oauth2_user",
			requestedLevel: expand.OAuth2User,
			f:              e.expandOAuth2User,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll},
			err:            utils.NewInvalidError(`invalid expand: "oauth2_user" must be "all" or empty`),
		}, {
			name:           "post",
			requestedLevel: expand.Post,
			f:              e.expandPost(&e.ExpandedContext.Post, e.UserAgentContext.PostID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "root_post",
			requestedLevel: expand.RootPost,
			f:              e.expandPost(&e.ExpandedContext.RootPost, e.UserAgentContext.RootPostID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "team_member",
			requestedLevel: expand.TeamMember,
			f:              e.expandTeamMember,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "team",
			requestedLevel: expand.Team,
			f:              e.expandTeam,
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "user",
			requestedLevel: expand.User,
			f:              e.expandUser(&e.ExpandedContext.User, e.UserID),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		},
	} {
		level, err := apps.ParseExpandLevel(step.requestedLevel, step.optionalByDefault)
		if err != nil {
			return nil, err
		}

		// Check the requested expand level to see if it's valid and if there is
		// anything to do.
		doExpand, err := func() (bool, error) {
			l := level.Level()
			if l == apps.ExpandNone {
				return false, nil
			}
			for _, expandable := range step.expandableAs {
				if l == expandable {
					return true, nil
				}
			}
			for _, known := range knownLevels {
				if l == known {
					return false, nil
				}
			}
			if step.err != nil {
				return false, errors.Wrap(step.err, "invalid expand level "+string(level))
			}
			return false, errors.New("invalid expand level " + string(level))
		}()
		if err != nil {
			return nil, err
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
	cc := e.Context
	cc.UserAgentContext.ChannelID = ""
	cc.UserAgentContext.TeamID = ""
	cc.UserAgentContext.RootPostID = ""
	cc.UserAgentContext.PostID = ""
	cc.UserID = ""
	cc.ActingUserID = ""

	return &cc, nil
}

func (e *expander) expandActingUserAccessToken(level apps.ExpandLevel) error {
	if !e.app.GrantedPermissions.Contains(apps.PermissionActAsUser) {
		return utils.NewForbiddenError("%s does not have permission to %s", e.app.AppID, apps.PermissionActAsUser)
	}

	token, err := e.r.UserAccessToken()
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
			return nil
		}

		switch level {
		case apps.ExpandID:
			*userPtr = &model.User{
				Id: userID,
			}

		case apps.ExpandSummary:
			user, err := e.client.GetUser(userID)
			if err != nil {
				return errors.Wrapf(err, "id: %s", userID)
			}
			*userPtr = &model.User{
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

		case apps.ExpandAll:
			user, err := e.client.GetUser(userID)
			if err != nil {
				return errors.Wrapf(err, "id: %s", userID)
			}
			user.Sanitize(map[string]bool{
				"email":    true,
				"fullname": true,
			})
			user.AuthData = nil
			*userPtr = user
		}
		return nil
	}
}

func (e *expander) expandApp(level apps.ExpandLevel) error {
	switch level {
	case apps.ExpandID:
		e.ExpandedContext.App = &apps.App{
			Manifest: apps.Manifest{
				AppID: e.app.AppID,
			},
			BotUserID: e.app.BotUserID,
		}
	case apps.ExpandAll, apps.ExpandSummary:
		e.ExpandedContext.App = &apps.App{
			Manifest: apps.Manifest{
				AppID:   e.app.AppID,
				Version: e.app.Version,
			},
			WebhookSecret: e.app.WebhookSecret,
			BotUserID:     e.app.BotUserID,
			BotUsername:   e.app.BotUsername,
		}
	}
	return nil
}

func (e *expander) expandChannelMember(level apps.ExpandLevel) error {
	channelID := e.UserAgentContext.ChannelID
	userID := e.ActingUserID
	if userID == "" {
		userID = e.UserID
	}
	if userID == "" || channelID == "" {
		return errors.New("no user ID or channel ID to expand")
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.ChannelMember = &model.ChannelMember{
			UserId:    userID,
			ChannelId: channelID,
		}
	case apps.ExpandSummary, apps.ExpandAll:
		cm, err := e.client.GetChannelMember(channelID, userID)
		if err != nil {
			return errors.Wrap(err, "failed to get channel membership")
		}
		e.ExpandedContext.ChannelMember = cm
	}
	return nil
}

func (e *expander) expandChannel(level apps.ExpandLevel) error {
	channelID := e.UserAgentContext.ChannelID
	if channelID == "" {
		return errors.New("no channel ID to expand")
	}

	switch level {
	case apps.ExpandDefault, apps.ExpandID:
		e.ExpandedContext.Channel = &model.Channel{
			Id: channelID,
		}

	case apps.ExpandSummary:
		channel, err := e.client.GetChannel(channelID)
		if err != nil {
			return errors.Wrap(err, "id: "+channelID)
		}
		e.ExpandedContext.Channel = &model.Channel{
			Id:          channel.Id,
			DeleteAt:    channel.DeleteAt,
			TeamId:      channel.TeamId,
			Type:        channel.Type,
			DisplayName: channel.DisplayName,
			Name:        channel.Name,
		}

	case apps.ExpandAll:
		channel, err := e.client.GetChannel(channelID)
		if err != nil {
			return errors.Wrap(err, "id: "+channelID)
		}
		e.ExpandedContext.Channel = channel
	}
	return nil
}

func (e *expander) expandTeam(level apps.ExpandLevel) error {
	teamID := e.UserAgentContext.TeamID
	if teamID == "" {
		return errors.New("no team ID to expand")
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.Team = &model.Team{
			Id: teamID,
		}
	case apps.ExpandSummary:
		team, err := e.client.GetTeam(teamID)
		if err != nil {
			return errors.Wrapf(err, "failed to get team %s", teamID)
		}
		e.ExpandedContext.Team = &model.Team{
			Id:          team.Id,
			DisplayName: team.DisplayName,
			Name:        team.Name,
			Description: team.Description,
			Email:       team.Email,
			Type:        team.Type,
		}

	case apps.ExpandAll:
		team, err := e.client.GetTeam(teamID)
		if err != nil {
			return errors.Wrapf(err, "failed to get team %s", teamID)
		}

		sanitized := *team
		sanitized.Sanitize()
		e.ExpandedContext.Team = &sanitized
	}

	return nil
}

func (e *expander) expandTeamMember(level apps.ExpandLevel) error {
	teamID := e.UserAgentContext.TeamID
	userID := e.ActingUserID
	if userID == "" {
		userID = e.UserID
	}
	if userID == "" || teamID == "" {
		return errors.New("no user ID or channel ID to expand")
	}

	switch level {
	case apps.ExpandID:
		e.ExpandedContext.TeamMember = &model.TeamMember{
			UserId: userID,
			TeamId: teamID,
		}
	case apps.ExpandSummary, apps.ExpandAll:
		tm, err := e.client.GetTeamMember(teamID, userID)
		if err != nil {
			return errors.Wrap(err, "failed to get team membership")
		}
		e.ExpandedContext.TeamMember = tm
	}
	return nil
}

func (e *expander) expandPost(postPtr **model.Post, postID string) expandFunc {
	return func(level apps.ExpandLevel) error {
		if postID == "" {
			return errors.New("no post ID to expand")
		}

		switch level {
		case apps.ExpandID:
			*postPtr = &model.Post{
				Id: postID,
			}
		case apps.ExpandSummary:
			post, err := e.client.GetPost(postID)
			if err != nil {
				return errors.Wrapf(err, "failed to get post %s", postID)
			}
			*postPtr = &model.Post{
				Id:        post.Id,
				Type:      post.Type,
				UserId:    post.UserId,
				ChannelId: post.ChannelId,
				RootId:    post.RootId,
				Message:   post.Message,
			}

		case apps.ExpandAll:
			post, err := e.client.GetPost(postID)
			if err != nil {
				return errors.Wrapf(err, "failed to get post %s", postID)
			}

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
		e.ExpandedContext.Locale = utils.GetLocale(e.r.Config().MattermostAPI(), e.r.Config().MattermostConfig().Config(), e.ActingUserID)
	}
	return nil
}

func (e *expander) expandOAuth2App(level apps.ExpandLevel) error {
	if !e.app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewForbiddenError("%s does not have permission to %s", e.app.AppID, apps.PermissionRemoteOAuth2)
	}

	conf := e.r.Config().Get()
	e.ExpandedContext.OAuth2.OAuth2App = e.app.RemoteOAuth2
	e.ExpandedContext.OAuth2.ConnectURL = conf.AppURL(e.app.AppID) + appspath.RemoteOAuth2Connect
	e.ExpandedContext.OAuth2.CompleteURL = conf.AppURL(e.app.AppID) + appspath.RemoteOAuth2Complete
	return nil
}

func (e *expander) expandOAuth2User(level apps.ExpandLevel) error {
	userID := e.ActingUserID
	if userID == "" {
		return errors.New("no acting user id to expand")
	}
	if !e.app.GrantedPermissions.Contains(apps.PermissionRemoteOAuth2) {
		return utils.NewForbiddenError("%s does not have permission to %s", e.app.AppID, apps.PermissionRemoteOAuth2)
	}

	data, err := e.appservices.GetOAuth2User(e.r, e.app.AppID, userID)
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

func (e *expander) consistencyCheck() error {
	if e.ExpandedContext.Post != nil {
		if e.ExpandedContext.Post.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded post's channel ID %s is different from user agent context %s",
				e.ExpandedContext.Post.ChannelId, e.UserAgentContext.ChannelID)
		}
	}

	if e.ExpandedContext.Channel != nil {
		if e.ExpandedContext.Channel.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded channel's team ID %s is different from user agent context %s",
				e.ExpandedContext.Channel.TeamId, e.UserAgentContext.TeamID)
		}
	}

	if e.ExpandedContext.ChannelMember != nil {
		if e.ExpandedContext.ChannelMember.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded channel member's channel ID %s is different from user agent context %s",
				e.ExpandedContext.ChannelMember.ChannelId, e.UserAgentContext.ChannelID)
		}
		if e.ExpandedContext.ChannelMember.UserId != e.ActingUserID && e.ExpandedContext.ChannelMember.UserId != e.UserID {
			return errors.Errorf("expanded channel member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	if e.ExpandedContext.TeamMember != nil {
		if e.ExpandedContext.TeamMember.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded team member's team ID %s is different from user agent context %s",
				e.ExpandedContext.TeamMember.TeamId, e.UserAgentContext.TeamID)
		}
		if e.ExpandedContext.TeamMember.UserId != e.ActingUserID && e.ExpandedContext.TeamMember.UserId != e.UserID {
			return errors.Errorf("expanded team member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	return nil
}
