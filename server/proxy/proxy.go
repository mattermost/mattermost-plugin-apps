// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Call(in Incoming, creq apps.CallRequest) apps.ProxyCallResponse {
	if creq.Context.AppID == "" {
		return apps.NewProxyCallResponse(
			apps.NewErrorCallResponse(
				utils.NewInvalidError("app_id is not set in Context, don't know what app to call")), nil)
	}

	app, err := p.store.App.Get(creq.Context.AppID)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), nil)
	}

	metadata := &apps.AppMetadataForClient{
		BotUserID:   app.BotUserID,
		BotUsername: app.BotUsername,
	}

	cresp := p.callApp(in, app, creq)
	return apps.NewProxyCallResponse(cresp, metadata)
}

func (p *Proxy) callApp(in Incoming, app *apps.App, creq apps.CallRequest) apps.CallResponse {
	if !p.appIsEnabled(app) {
		return apps.NewErrorCallResponse(errors.Errorf("%s is disabled", app.AppID))
	}

	if creq.Path[0] != '/' {
		return apps.NewErrorCallResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	cc := creq.Context
	cc = in.updateContext(cc)
	creq.Context, err = p.expandContext(in, app, &cc, creq.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	cresp := upstream.Call(up, *app, creq)
	if cresp.Type == "" {
		cresp.Type = apps.CallResponseTypeOK
	}

	if cresp.Form != nil {
		if cresp.Form.Icon != "" {
			conf, _, log := p.conf.Basic()
			log = log.With("app_id", app.AppID)
			icon, err := normalizeStaticPath(conf, cc.AppID, cresp.Form.Icon)
			if err != nil {
				log.WithError(err).Debugw("Invalid icon path in form. Ignoring it.", "icon", cresp.Form.Icon)
				cresp.Form.Icon = ""
			} else {
				cresp.Form.Icon = icon
			}
			clean, problems := cleanForm(*cresp.Form)
			for _, prob := range problems {
				log.WithError(prob).Debugw("invalid form")
			}
			cresp.Form = &clean
		}
	}

	return cresp
}

// normalizeStaticPath converts a given URL to a absolute one pointing to a static asset if needed.
// If icon is an absolute URL, it's not changed.
// Otherwise assume it's a path to a static asset and the static path URL prepended.
func normalizeStaticPath(conf config.Config, appID apps.AppID, icon string) (string, error) {
	if !strings.HasPrefix(icon, "http://") && !strings.HasPrefix(icon, "https://") {
		cleanIcon, err := utils.CleanStaticPath(icon)
		if err != nil {
			return "", errors.Wrap(err, "invalid icon path")
		}

		icon = conf.StaticURL(appID, cleanIcon)
	}

	return icon, nil
}

func (p *Proxy) Notify(base apps.Context, subj apps.Subject) error {
	subs, err := p.store.Subscription.Get(subj, base.TeamID, base.ChannelID)
	if err != nil {
		return err
	}

	return p.notify(base, subs)
}

func (p *Proxy) notify(base apps.Context, subs []apps.Subscription) error {
	for _, sub := range subs {
		err := p.notifyForSubscription(&base, sub)
		if err != nil {
			p.conf.Logger().WithError(err).Debugw("Error sending subscription notification to app",
				"app_id", sub.AppID,
				"subject", sub.Subject)
		}
	}

	return nil
}

func (p *Proxy) notifyForSubscription(base *apps.Context, sub apps.Subscription) error {
	creq := apps.CallRequest{
		Call: sub.Call,
	}
	app, err := p.store.App.Get(sub.AppID)
	if err != nil {
		return err
	}
	if !p.appIsEnabled(app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}

	creq.Context, err = p.expandContext(Incoming{}, app, base, sub.Call.Expand)
	if err != nil {
		return err
	}
	creq.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(app)
	if err != nil {
		return err
	}
	return upstream.Notify(up, *app, creq)
}

func (p *Proxy) NotifyRemoteWebhook(app apps.App, data []byte, webhookPath string) error {
	if !p.appIsEnabled(&app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	up, err := p.upstreamForApp(&app)
	if err != nil {
		return err
	}

	var datav interface{}
	err = json.Unmarshal(data, &datav)
	if err != nil {
		// if the data can not be decoded as JSON, send it "as is", as a string.
		datav = string(data)
	}

	conf := p.conf.Get()
	cc := contextForApp(&app, apps.Context{}, conf)
	// Set acting user to bot.
	cc.ActingUserID = app.BotUserID
	cc.ActingUserAccessToken = app.BotAccessToken

	// TODO: do we need to customize the Expand & State for the webhook Call?
	return upstream.Notify(up, app, apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(apps.PathWebhook, webhookPath),
		},
		Context: cc,
		Values: map[string]interface{}{
			"data": datav,
		},
	})
}

func (p *Proxy) NotifyMessageHasBeenPosted(post *model.Post, cc apps.Context) error {
	postSubs, err := p.store.Subscription.Get(apps.SubjectPostCreated, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrap(err, "failed to get post_created subscriptions")
	}

	subs := append([]apps.Subscription{}, postSubs...)
	mentions := model.PossibleAtMentions(post.Message)

	botCanRead := map[string]bool{}
	if len(mentions) > 0 {
		appsMap := p.store.App.AsMap()
		mentionSubs, err := p.store.Subscription.Get(apps.SubjectBotMentioned, cc.TeamID, cc.ChannelID)
		if err != nil && err != utils.ErrNotFound {
			return errors.Wrap(err, "failed to get bot_mentioned subscriptions")
		}

		for _, sub := range mentionSubs {
			app, ok := appsMap[sub.AppID]
			if !ok {
				continue
			}
			for _, mention := range mentions {
				if mention == app.BotUsername {
					_, ok := botCanRead[app.BotUserID]
					if ok {
						// already processed this bot for this post
						continue
					}

					canRead := p.conf.MattermostAPI().User.HasPermissionToChannel(app.BotUserID, post.ChannelId, model.PERMISSION_READ_CHANNEL)
					botCanRead[app.BotUserID] = canRead

					if canRead {
						subs = append(subs, sub)
					}
				}
			}
		}
	}

	if len(subs) == 0 {
		return nil
	}

	return p.notify(cc, subs)
}

func (p *Proxy) NotifyUserHasJoinedChannel(cc apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserJoinedChannel, apps.SubjectBotJoinedChannel)
}

func (p *Proxy) NotifyUserHasLeftChannel(cc apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserLeftChannel, apps.SubjectBotLeftChannel)
}

func (p *Proxy) NotifyUserHasJoinedTeam(cc apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserJoinedTeam, apps.SubjectBotJoinedTeam)
}

func (p *Proxy) NotifyUserHasLeftTeam(cc apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserLeftTeam, apps.SubjectBotLeftTeam)
}

func (p *Proxy) notifyJoinLeave(cc apps.Context, subject, botSubject apps.Subject) error {
	userSubs, err := p.store.Subscription.Get(subject, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrapf(err, "failed to get %s subscriptions", subject)
	}

	botSubs, err := p.store.Subscription.Get(botSubject, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrapf(err, "failed to get %s subscriptions", botSubject)
	}

	subs := []apps.Subscription{}
	subs = append(subs, userSubs...)

	appsMap := p.store.App.AsMap()
	for _, sub := range botSubs {
		app, ok := appsMap[sub.AppID]
		if !ok {
			continue
		}

		if app.BotUserID == cc.UserID {
			subs = append(subs, sub)
		}
	}

	return p.notify(cc, subs)
}

func (p *Proxy) GetStatic(appID apps.AppID, path string) (io.ReadCloser, int, error) {
	app, err := p.store.App.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		return nil, status, err
	}

	return p.getStatic(app, path)
}

func (p *Proxy) getStatic(app *apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return up.GetStatic(*app, path)
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
	if app.AppType == apps.AppTypeBuiltin {
		u, ok := p.builtinUpstreams[app.AppID]
		if !ok {
			return nil, errors.Wrapf(utils.ErrNotFound, "no builtin %s", app.AppID)
		}
		return u, nil
	}

	conf := p.conf.Get()
	err := isAppTypeSupported(conf, app.AppType)
	if err != nil {
		return nil, err
	}

	upv, ok := p.upstreams.Load(app.AppType)
	if !ok {
		return nil, utils.NewInvalidError("invalid app type: %s", app.AppType)
	}
	up, ok := upv.(upstream.Upstream)
	if !ok {
		return nil, utils.NewInvalidError("invalid Upstream for: %s", app.AppType)
	}
	return up, nil
}

func isAppTypeSupported(conf config.Config, appType apps.AppType) error {
	supportedTypes := []apps.AppType{
		apps.AppTypeAWSLambda,
		apps.AppTypeBuiltin,
		apps.AppTypePlugin,
	}
	mode := "Mattermost Cloud"

	switch {
	case conf.DeveloperMode:
		return nil

	case conf.MattermostCloudMode:

	case !conf.MattermostCloudMode:
		// Self-managed
		supportedTypes = append(supportedTypes, apps.AppTypeHTTP)
		mode = "Self-managed"

	default:
		return errors.New("unreachable")
	}

	for _, t := range supportedTypes {
		if appType == t {
			return nil
		}
	}
	return utils.NewForbiddenError("%s is not allowed in %s mode, only %s", appType, mode, supportedTypes)
}
