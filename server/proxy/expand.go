package proxy

import (
	"encoding/json"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
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
	conf   config.Config

	botAccessToken        string
	actingUserAccessToken string
}

// expandContext performs `expand` on a `context` of an outgoing CallRequest. It
// ensures a fresh/correct ExpandedContext. It preserves the UserAgentContext
// since it is used during expand, which cleans it out before returning.
//
// The request `r` configured with correct source and destination app IDs, as
// well as the acting user data. (destination) `app` is passed in as a shortcut
// since it's already available in all callers.
func (p *Proxy) expandContext(r *incoming.Request, app *apps.App, cc *apps.Context, expand *apps.Expand) (_ *apps.Context, err error) {
	conf := r.Config().Get()
	defer func() {
		if err != nil && conf.DeveloperMode {
			r.Log.WithError(err).Debugw("Expand failed")
		}
	}()

	if r.Destination() == "" {
		return nil, errors.New("missing destination app ID in request")
	}
	if cc == nil {
		cc = &apps.Context{}
	}

	e := &expander{
		conf:    conf,
		app:     app,
		Context: *cc,
		proxy:   p,
		r:       r,
	}

	e.ExpandedContext = apps.ExpandedContext{
		AppPath:           path.Join(conf.PluginURLPath, appspath.Apps, string(app.AppID)),
		BotUserID:         app.BotUserID,
		DeveloperMode:     conf.DeveloperMode,
		MattermostSiteURL: conf.MattermostSiteURL,
	}

	if app.GrantedPermissions.Contains(apps.PermissionActAsBot) && app.BotUserID != "" {
		botAccessToken, err := e.getBotAccessToken()
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
	}{
		{
			name:           "acting_user_access_token",
			requestedLevel: expand.ActingUserAccessToken,
			f:              e.expandActingUserAccessToken,
			expandableAs:   []apps.ExpandLevel{apps.ExpandAll, apps.ExpandSummary},
		}, {
			name:           "acting_user",
			requestedLevel: expand.ActingUser,
			f:              e.expandUser(&e.ExpandedContext.ActingUser, e.r.ActingUserID()),
			expandableAs:   []apps.ExpandLevel{apps.ExpandID, apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "app",
			requestedLevel: expand.App,
			f:              e.expandApp,
			expandableAs:   []apps.ExpandLevel{apps.ExpandSummary, apps.ExpandAll},
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
			expandableAs:   []apps.ExpandLevel{apps.ExpandSummary, apps.ExpandAll},
		}, {
			name:           "oauth2_user",
			requestedLevel: expand.OAuth2User,
			f:              e.expandOAuth2User,
			expandableAs:   []apps.ExpandLevel{apps.ExpandSummary apps.ExpandAll},
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
		required, level, err := apps.ParseExpandLevel(step.requestedLevel)
		if err != nil {
			return nil, err
		}

		// Check the requested expand level to see if it's valid and if there is
		// anything to do.
		doExpand := false
		for _, expandable := range step.expandableAs {
			if level == expandable {
				doExpand = true
				break
			}
		}
		if !doExpand {
			continue
		}

		// Don't attempt to make a client (session and all) unless we need to expand.
		err = e.ensureClient()
		if err != nil {
			return nil, err
		}

		// Execute the expand function, skip the error unless the field is
		// required.
		if err := step.f(level); err != nil {
			if e.conf.DeveloperMode {
				e.r.Log.WithError(err).Debugf("failed to expand field %s", step.name)
			}
			if required {
				return nil, errors.Wrap(err, "failed to expand required "+step.name)
			}
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

	token, err := e.getActingUserAccessToken()
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

		user, err := e.client.GetUser(userID)
		if err != nil {
			return errors.Wrapf(err, "id: %s", userID)
		}
		*userPtr = apps.StripUser(user, level)
		return nil
	}
}

func (e *expander) expandApp(level apps.ExpandLevel) error {
	e.ExpandedContext.App = e.app.Strip(level)

	e.ExpandedContext.App.WebhookSecret = ""
	if level == apps.ExpandAll && e.r.RequireSysadminOrPlugin() == nil {
		e.ExpandedContext.App.WebhookSecret = e.app.WebhookSecret
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
	e.ExpandedContext.ChannelMember = apps.StripChannelMember(cm, level)
	return nil
}

func (e *expander) expandChannel(level apps.ExpandLevel) error {
	channelID := e.UserAgentContext.ChannelID
	if channelID == "" {
		return errors.New("no channel ID to expand")
	}

	channel, err := e.client.GetChannel(channelID)
	if err != nil {
		if level == apps.ExpandID {
			// Always expand Channel and Team IDs to make `bot_left_channel`
			// work. This really should be fixed on the server by redefining the
			// semantics to UserWillLeaveChannel, called before the user's
			// permission disappear.
			e.ExpandedContext.Channel = &model.Channel{
				Id:     channelID,
				TeamId: e.UserAgentContext.TeamID,
			}
			return nil
		}
		return errors.Wrap(err, "id: "+channelID)
	}
	e.ExpandedContext.Channel = apps.StripChannel(channel, level)
	return nil
}

func (e *expander) expandTeam(level apps.ExpandLevel) error {
	teamID := e.UserAgentContext.TeamID
	if teamID == "" {
		return errors.New("no team ID to expand")
	}

	team, err := e.client.GetTeam(teamID)
	if err != nil {
		if level == apps.ExpandID {
			// Always expand Team ID to make `bot_left_team` works. This really
			// should be fixed on the server by redefining the semantics to
			// UserWillLeaveTeam, called before the user's permission disappear.
			e.ExpandedContext.Team = &model.Team{
				Id: teamID,
			}
			return nil
		}
		return errors.Wrapf(err, "failed to get team %s", teamID)
	}
	e.ExpandedContext.Team = apps.StripTeam(team, level)
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
	e.ExpandedContext.TeamMember = apps.StripTeamMember(tm, level)
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
		*postPtr = apps.StripPost(post, level)
		return nil
	}
}

func (e *expander) expandLocale(level apps.ExpandLevel) error {
	confService := e.r.Config()
	if e.ExpandedContext.ActingUser != nil {
		e.ExpandedContext.Locale = utils.GetLocaleWithUser(confService.MattermostConfig().Config(), e.ExpandedContext.ActingUser)
	} else {
		e.ExpandedContext.Locale = utils.GetLocale(confService.MattermostAPI(), confService.MattermostConfig().Config(), e.r.ActingUserID())
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

	e.ExpandedContext.OAuth2.OAuth2App = to.RemoteOAuth2
	e.ExpandedContext.OAuth2.ConnectURL = e.conf.AppURL(to.AppID) + appspath.RemoteOAuth2Connect
	e.ExpandedContext.OAuth2.CompleteURL = e.conf.AppURL(to.AppID) + appspath.RemoteOAuth2Complete
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

func (e *expander) ensureClient() error {
	client := e.client
	if client != nil {
		return nil
	}
	if e.proxy.expandClientOverride != nil {
		e.client = e.proxy.expandClientOverride
		return nil
	}
	app, err := e.proxy.getEnabledDestination(e.r)
	if err != nil {
		return err
	}
	conf := e.proxy.conf.Get()

	switch {
	case app.DeployType == apps.DeployBuiltin:
		client = mmclient.NewRPCClient(e.proxy.conf.MattermostAPI())

	case app.GrantedPermissions.Contains(apps.PermissionActAsUser) && e.r.ActingUserID() != "":
		token, err := e.getActingUserAccessToken()
		if err != nil {
			return errors.Wrap(err, "failed to get the current user's access token")
		}
		client = mmclient.NewHTTPClient(e.proxy.conf.Get(), token)

	case app.GrantedPermissions.Contains(apps.PermissionActAsBot):
		accessToken, err := e.getBotAccessToken()
		if err != nil {
			return errors.Wrap(err, "failed to get the bot's access token")
		}
		client = mmclient.NewHTTPClient(conf, accessToken)

	default:
		return utils.NewUnauthorizedError("apps without any ActAs* permission can't expand")
	}
	e.client = client
	return nil
}

func (e *expander) getBotAccessToken() (string, error) {
	if e.botAccessToken != "" {
		return e.botAccessToken, nil
	}
	session, err := e.proxy.sessionService.GetOrCreate(e.r, e.app.BotUserID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get bot session")
	}
	e.botAccessToken = session.Token
	return session.Token, nil
}

func (e *expander) getActingUserAccessToken() (string, error) {
	if e.actingUserAccessToken != "" {
		return e.actingUserAccessToken, nil
	}
	session, err := e.proxy.sessionService.GetOrCreate(e.r, e.r.ActingUserID())
	if err != nil {
		return "", errors.Wrap(err, "failed to get session")
	}
	e.actingUserAccessToken = session.Token
	return session.Token, nil
}

func (e *expander) consistencyCheck() error {
	if e.ExpandedContext.Post != nil {
		if e.UserAgentContext.ChannelID != "" && e.ExpandedContext.Post.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded post's channel ID %s is different from user agent context %s",
				e.ExpandedContext.Post.ChannelId, e.UserAgentContext.ChannelID)
		}
	}

	if e.ExpandedContext.Channel != nil {
		if e.UserAgentContext.TeamID != "" && e.ExpandedContext.Channel.Type != model.ChannelTypeDirect && e.ExpandedContext.Channel.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded channel's team ID %s is different from user agent context %s",
				e.ExpandedContext.Channel.TeamId, e.UserAgentContext.TeamID)
		}
	}

	if e.ExpandedContext.ChannelMember != nil {
		if e.UserAgentContext.ChannelID != "" && e.ExpandedContext.ChannelMember.ChannelId != e.UserAgentContext.ChannelID {
			return errors.Errorf("expanded channel member's channel ID %s is different from user agent context %s",
				e.ExpandedContext.ChannelMember.ChannelId, e.UserAgentContext.ChannelID)
		}
		if e.UserID != "" && e.ExpandedContext.ChannelMember.UserId != e.r.ActingUserID() && e.ExpandedContext.ChannelMember.UserId != e.UserID {
			return errors.Errorf("expanded channel member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	if e.ExpandedContext.TeamMember != nil {
		if e.UserAgentContext.TeamID != "" && e.ExpandedContext.TeamMember.TeamId != e.UserAgentContext.TeamID {
			return errors.Errorf("expanded team member's team ID %s is different from user agent context %s",
				e.ExpandedContext.TeamMember.TeamId, e.UserAgentContext.TeamID)
		}
		if e.UserID != "" && e.ExpandedContext.TeamMember.UserId != e.r.ActingUserID() && e.ExpandedContext.TeamMember.UserId != e.UserID {
			return errors.Errorf("expanded team member's user ID %s is different from user agent context",
				e.ExpandedContext.ChannelMember.UserId)
		}
	}

	return nil
}
