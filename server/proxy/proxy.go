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

	creq.Context.Locale = utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), creq.Context.ActingUserID)

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

	p.cleanForm(callResponse.Form)

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

// cleanForm removes:
// - Fields without a name
// - Fields with labels (either natural or defaulted from names) with more than one word
// - Fields that have the same label as previous fields
// - Invalid select static fields and their invalid options
func (p *Proxy) cleanForm(form *apps.Form) {
	if form == nil {
		return
	}

	toRemove := []int{}
	usedLabels := map[string]bool{}
	for i, field := range form.Fields {
		if field.Name == "" {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App from malformed: Field with no name", "field", field)
			continue
		}
		if strings.ContainsAny(field.Name, " \t") {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App form malformed: Name must be a single word", "name", field.Name)
			continue
		}

		label := field.Label
		if label == "" {
			label = field.Name
		}
		if strings.ContainsAny(label, " \t") {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App form malformed: Label must be a single word", "label", label)
			continue
		}

		if usedLabels[label] {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App from malformed: Field label repeated. Only getting first field with the label.", "label", label)
			continue
		}

		if field.Type == apps.FieldTypeStaticSelect {
			p.cleanStaticSelect(field)
			if len(field.SelectStaticOptions) == 0 {
				toRemove = append([]int{i}, toRemove...)
				p.conf.MattermostAPI().Log.Debug("App from malformed: Static field without opions.", "label", label)
				continue
			}
		}

		usedLabels[label] = true
	}

	for _, i := range toRemove {
		form.Fields = append(form.Fields[:i], form.Fields[i+1:]...)
	}
}

// cleanStaticSelect removes:
// - Options with empty label (either natural or defaulted form the value)
// - Options that have the same label as the previous options
// - Options that have the same value as the previous options
func (p *Proxy) cleanStaticSelect(field *apps.Field) {
	toRemove := []int{}
	usedLabels := map[string]bool{}
	usedValues := map[string]bool{}
	for i, option := range field.SelectStaticOptions {
		label := option.Label
		if label == "" {
			label = option.Value
		}

		if label == "" {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App from malformed: Option with no label", "field", field, "option value", option.Value)
			continue
		}

		if usedLabels[label] {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App from malformed: Repeated label on select option. Only getting first value with the label", "field", field, "option", option)
			continue
		}

		if usedValues[option.Value] {
			toRemove = append([]int{i}, toRemove...)
			p.conf.MattermostAPI().Log.Debug("App from malformed: Repeated value on select option. Only getting first value with the value", "field", field, "option", option)
			continue
		}

		usedLabels[label] = true
		usedValues[option.Value] = true
	}

	for _, i := range toRemove {
		field.SelectStaticOptions = append(field.SelectStaticOptions[:i], field.SelectStaticOptions[i+1:]...)
	}
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
			log := p.conf.Logger().WithError(err).With("app_id", sub.AppID, "subject", sub.Subject)
			log.Debugf("Error sending subscription notification to app")
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
			Locale:       utils.GetLocale(p.conf.MattermostAPI(), p.conf.MattermostConfig().Config(), app.BotUserID),
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
