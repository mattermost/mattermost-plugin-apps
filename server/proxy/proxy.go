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

func (p *Proxy) Call(sessionID, actingUserID string, creq *apps.CallRequest) *apps.ProxyCallResponse {
	if creq.Context == nil || creq.Context.AppID == "" {
		return apps.NewProxyCallResponse(
			apps.NewErrorCallResponse(
				utils.NewInvalidError("must provide Context and set the app ID")), nil)
	}

	app, err := p.store.App.Get(creq.Context.AppID)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), nil)
	}

	metadata := &apps.AppMetadataForClient{
		BotUserID:   app.BotUserID,
		BotUsername: app.BotUsername,
	}

	cresp := p.callApp(app, sessionID, actingUserID, creq)
	return apps.NewProxyCallResponse(cresp, metadata)
}

func (p *Proxy) callApp(app *apps.App, sessionID, actingUserID string, creq *apps.CallRequest) *apps.CallResponse {
	if !p.AppIsEnabled(app) {
		return apps.NewErrorCallResponse(errors.Errorf("%s is disabled", app.AppID))
	}

	if actingUserID != "" {
		creq.Context.ActingUserID = actingUserID
		creq.Context.UserID = actingUserID
	}

	if creq.Path[0] != '/' {
		return apps.NewErrorCallResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}

	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(&app.Manifest)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}

	// Clear any ExpandedContext as it should always be set by an expander for security reasons
	creq.Context.ExpandedContext = apps.ExpandedContext{}

	conf, _, log := p.conf.Basic()
	cc := conf.SetContextDefaultsForApp(creq.Context.AppID, creq.Context)

	expander := p.newExpander(cc, p.conf, p.store, sessionID)
	cc, err = expander.ExpandForApp(app, creq.Expand)
	if err != nil {
		return apps.NewErrorCallResponse(err)
	}
	clone := *creq
	clone.Context = cc

	callResponse := upstream.Call(up, app, &clone)

	if callResponse.Type == "" {
		callResponse.Type = apps.CallResponseTypeOK
	}

	if callResponse.Form != nil {
		if callResponse.Form.Icon != "" {
			icon, err := normalizeStaticPath(conf, cc.AppID, callResponse.Form.Icon)
			if err != nil {
				log.WithError(err).Debugw("Invalid icon path in form. Ignoring it.",
					"app_id", app.AppID,
					"icon", callResponse.Form.Icon)
				callResponse.Form.Icon = ""
			} else {
				callResponse.Form.Icon = icon
			}
			clean, problems := cleanForm(*callResponse.Form)
			for _, prob := range problems {
				log.WithError(prob).Debugf("invalid form field in bingding")
			}
			callResponse.Form = &clean
		}
	}

	return callResponse
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

func (p *Proxy) Notify(cc *apps.Context, subj apps.Subject) error {
	subs, err := p.store.Subscription.Get(subj, cc.TeamID, cc.ChannelID)
	if err != nil {
		return err
	}

	return p.notify(cc, subs)
}

func (p *Proxy) notify(cc *apps.Context, subs []*apps.Subscription) error {
	expander := p.newExpander(cc, p.conf, p.store, "")

	for _, sub := range subs {
		err := p.notifyForSubscription(cc, expander, sub)
		if err != nil {
			p.conf.Logger().WithError(err).Debugw("Error sending subscription notification to app",
				"app_id", sub.AppID,
				"subject", sub.Subject)
		}
	}

	return nil
}

func (p *Proxy) notifyForSubscription(cc *apps.Context, expander *expander, sub *apps.Subscription) error {
	call := sub.Call
	if call == nil {
		return errors.New("nothing to call")
	}

	callRequest := &apps.CallRequest{Call: *call}
	app, err := p.store.App.Get(sub.AppID)
	if err != nil {
		return err
	}
	if !p.AppIsEnabled(app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}

	callRequest.Context, err = expander.ExpandForApp(app, callRequest.Expand)
	if err != nil {
		return err
	}
	callRequest.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(&app.Manifest)
	if err != nil {
		return err
	}
	return upstream.Notify(up, app, callRequest)
}

func (p *Proxy) NotifyRemoteWebhook(app *apps.App, data []byte, webhookPath string) error {
	if !p.AppIsEnabled(app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	up, err := p.upstreamForApp(&app.Manifest)
	if err != nil {
		return err
	}

	var datav interface{}
	err = json.Unmarshal(data, &datav)
	if err != nil {
		// if the data can not be decoded as JSON, send it "as is", as a string.
		datav = string(data)
	}

	// TODO: do we need to customize the Expand & State for the webhook Call?
	creq := &apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(apps.PathWebhook, webhookPath),
		},
		Context: p.conf.Get().SetContextDefaultsForApp(app.AppID, &apps.Context{
			ActingUserID: app.BotUserID,
		}),
		Values: map[string]interface{}{
			"data": datav,
		},
	}
	expander := p.newExpander(creq.Context, p.conf, p.store, "")
	creq.Context, err = expander.ExpandForApp(app, creq.Expand)
	if err != nil {
		return err
	}

	return upstream.Notify(up, app, creq)
}

func (p *Proxy) NotifyMessageHasBeenPosted(post *model.Post, cc *apps.Context) error {
	postSubs, err := p.store.Subscription.Get(apps.SubjectPostCreated, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrap(err, "failed to get post_created subscriptions")
	}

	subs := []*apps.Subscription{}
	subs = append(subs, postSubs...)
	mentions := model.PossibleAtMentions(post.Message)

	botCanRead := map[string]bool{}
	if len(mentions) > 0 {
		appsMap := p.store.App.AsMap()
		mentionSubs, err := p.store.Subscription.Get(apps.SubjectBotMentioned, cc.TeamID, cc.ChannelID)
		if err != nil && err != utils.ErrNotFound {
			return errors.Wrap(err, "failed to get bot_mentioned subscriptions")
		}

		for _, sub := range mentionSubs {
			app := appsMap[sub.AppID]
			if app == nil {
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

func (p *Proxy) NotifyUserHasJoinedChannel(cc *apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserJoinedChannel, apps.SubjectBotJoinedChannel)
}

func (p *Proxy) NotifyUserHasLeftChannel(cc *apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserLeftChannel, apps.SubjectBotLeftChannel)
}

func (p *Proxy) NotifyUserHasJoinedTeam(cc *apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserJoinedTeam, apps.SubjectBotJoinedTeam)
}

func (p *Proxy) NotifyUserHasLeftTeam(cc *apps.Context) error {
	return p.notifyJoinLeave(cc, apps.SubjectUserLeftTeam, apps.SubjectBotLeftTeam)
}

func (p *Proxy) notifyJoinLeave(cc *apps.Context, subject, botSubject apps.Subject) error {
	userSubs, err := p.store.Subscription.Get(subject, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrapf(err, "failed to get %s subscriptions", subject)
	}

	botSubs, err := p.store.Subscription.Get(botSubject, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrapf(err, "failed to get %s subscriptions", botSubject)
	}

	subs := []*apps.Subscription{}
	subs = append(subs, userSubs...)

	appsMap := p.store.App.AsMap()
	for _, sub := range botSubs {
		app := appsMap[sub.AppID]
		if app == nil {
			continue
		}

		if app.BotUserID == cc.UserID {
			subs = append(subs, sub)
		}
	}

	return p.notify(cc, subs)
}

func (p *Proxy) GetStatic(appID apps.AppID, path string) (io.ReadCloser, int, error) {
	m, err := p.store.Manifest.Get(appID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, utils.ErrNotFound) {
			status = http.StatusNotFound
		}
		return nil, status, err
	}

	return p.getStatic(m, path)
}

func (p *Proxy) getStatic(m *apps.Manifest, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(m)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(m, path)
}

func (p *Proxy) upstreamForApp(m *apps.Manifest) (upstream.Upstream, error) {
	if m.AppType == apps.AppTypeBuiltin {
		u, ok := p.builtinUpstreams[m.AppID]
		if !ok {
			return nil, errors.Wrapf(utils.ErrNotFound, "no builtin %s", m.AppID)
		}
		return u, nil
	}

	conf := p.conf.Get()
	err := isAppTypeSupported(conf, m.AppType)
	if err != nil {
		return nil, err
	}

	upv, ok := p.upstreams.Load(m.AppType)
	if !ok {
		return nil, utils.NewInvalidError("invalid app type: %s", m.AppType)
	}
	up, ok := upv.(upstream.Upstream)
	if !ok {
		return nil, utils.NewInvalidError("invalid Upstream for: %s", m.AppType)
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
