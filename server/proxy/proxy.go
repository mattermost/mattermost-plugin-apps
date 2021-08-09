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
	"github.com/mattermost/mattermost-plugin-apps/upstream/upaws"
	"github.com/mattermost/mattermost-plugin-apps/upstream/uphttp"
	"github.com/mattermost/mattermost-plugin-apps/upstream/upplugin"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Call(sessionID, actingUserID string, creq *apps.CallRequest) *apps.ProxyCallResponse {
	conf, _, log := p.conf.Basic()

	if creq.Context == nil || creq.Context.AppID == "" {
		resp := apps.NewErrorCallResponse(utils.NewInvalidError("must provide Context and set the app ID"))
		return apps.NewProxyCallResponse(resp, nil)
	}

	if actingUserID != "" {
		creq.Context.ActingUserID = actingUserID
		creq.Context.UserID = actingUserID
	}

	app, err := p.store.App.Get(creq.Context.AppID)

	var metadata *apps.AppMetadataForClient
	if app != nil {
		metadata = &apps.AppMetadataForClient{
			BotUserID:   app.BotUserID,
			BotUsername: app.BotUsername,
		}
	}

	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}

	if creq.Path[0] != '/' {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path)), metadata)
	}

	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}

	// Clear any ExpandedContext as it should always be set by an expander for security reasons
	creq.Context.ExpandedContext = apps.ExpandedContext{}

	cc := conf.SetContextDefaultsForApp(creq.Context.AppID, creq.Context)

	expander := p.newExpander(cc, p.conf, p.store, sessionID)
	cc, err = expander.ExpandForApp(app, creq.Expand)
	if err != nil {
		return apps.NewProxyCallResponse(apps.NewErrorCallResponse(err), metadata)
	}
	clone := *creq
	clone.Context = cc

	callResponse := upstream.Call(up, &clone)

	if callResponse.Type == "" {
		callResponse.Type = apps.CallResponseTypeOK
	}

	if callResponse.Form != nil && callResponse.Form.Icon != "" {
		icon, err := normalizeStaticPath(conf, cc.AppID, callResponse.Form.Icon)
		if err != nil {
			log.WithError(err).Debugw("Invalid icon path in form. Ignoring it.",
				"app_id", app.AppID,
				"icon", callResponse.Form.Icon)
			callResponse.Form.Icon = ""
		} else {
			callResponse.Form.Icon = icon
		}
	}

	return apps.NewProxyCallResponse(callResponse, metadata)
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
			p.mm.Log.Debug("Error sending subscription notification to app", "app_id", sub.AppID, "subject", sub.Subject, "err", err.Error())
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
	callRequest.Context, err = expander.ExpandForApp(app, callRequest.Expand)
	if err != nil {
		return err
	}
	callRequest.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(app)
	if err != nil {
		return err
	}
	return upstream.Notify(up, callRequest)
}

func (p *Proxy) NotifyRemoteWebhook(app *apps.App, data []byte, webhookPath string) error {
	if !app.GrantedPermissions.Contains(apps.PermissionRemoteWebhooks) {
		return utils.NewForbiddenError("%s does not have permission %s", app.AppID, apps.PermissionRemoteWebhooks)
	}

	up, err := p.upstreamForApp(app)
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

	return upstream.Notify(up, creq)
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

					canRead := p.mm.User.HasPermissionToChannel(app.BotUserID, post.ChannelId, model.PERMISSION_READ_CHANNEL)
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
	up, err := p.staticUpstreamForManifest(m)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(path)
}

func (p *Proxy) getStaticForApp(app *apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.staticUpstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return up.GetStatic(path)
}

func (p *Proxy) staticUpstreamForManifest(m *apps.Manifest) (upstream.StaticUpstream, error) {
	switch m.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewStaticUpstream(m, p.httpOut), nil

	case apps.AppTypeAWSLambda:
		return upaws.NewStaticUpstream(m, p.aws, p.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		return nil, errors.New("static assets are not supported for builtin apps")

	case apps.AppTypePlugin:
		app, err := p.store.App.Get(m.AppID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get app for static asset")
		}
		return upplugin.NewStaticUpstream(app, &p.conf.MattermostAPI().Plugin), nil

	default:
		return nil, utils.NewInvalidError("not a valid app type: %s", m.AppType)
	}
}

func (p *Proxy) staticUpstreamForApp(app *apps.App) (upstream.StaticUpstream, error) {
	switch app.AppType {
	case apps.AppTypeHTTP, apps.AppTypeAWSLambda, apps.AppTypeBuiltin:
		return p.staticUpstreamForManifest(&app.Manifest)

	case apps.AppTypePlugin:
		return upplugin.NewStaticUpstream(app, &p.conf.MattermostAPI().Plugin), nil

	default:
		return nil, utils.NewInvalidError("not a valid app type: %s", app.AppType)
	}
}

func (p *Proxy) upstreamForApp(app *apps.App) (upstream.Upstream, error) {
	conf, mm, _ := p.conf.Basic()
	if !p.AppIsEnabled(app) {
		return nil, errors.Errorf("%s is disabled", app.AppID)
	}
	err := isAppTypeSupported(conf, &app.Manifest)
	if err != nil {
		return nil, err
	}

	switch app.AppType {
	case apps.AppTypeHTTP:
		return uphttp.NewUpstream(app, p.httpOut), nil

	case apps.AppTypeAWSLambda:
		return upaws.NewUpstream(app, p.aws, p.s3AssetBucket), nil

	case apps.AppTypeBuiltin:
		up := p.builtinUpstreams[app.AppID]
		if up == nil {
			return nil, utils.NewNotFoundError("builtin app not found: %s", app.AppID)
		}
		return up, nil
	case apps.AppTypePlugin:
		return upplugin.NewUpstream(app, &mm.Plugin), nil
	default:
		return nil, utils.NewInvalidError("invalid app type: %s", app.AppType)
	}
}

func isAppTypeSupported(conf config.Config, m *apps.Manifest) error {
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
		if m.AppType == t {
			return nil
		}
	}
	return utils.NewForbiddenError("%s is not allowed in %s mode, only %s", m.AppType, mode, supportedTypes)
}
