// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	appspath "github.com/mattermost/mattermost-plugin-apps/apps/path"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

// CallResponse contains everything the CallResponse struct contains, plus some additional
// data for the client, such as information about the App's bot account.
//
// Apps will use the CallResponse struct to respond to a CallRequest, and the proxy will
// decorate the response using the CallResponse to provide additional information.
type CallResponse struct {
	apps.CallResponse

	// Used to provide info about the App to client, e.g. the bot user id
	AppMetadata AppMetadataForClient `json:"app_metadata"`
}

type AppMetadataForClient struct {
	BotUserID   string `json:"bot_user_id,omitempty"`
	BotUsername string `json:"bot_username,omitempty"`
}

func NewProxyCallResponse(response apps.CallResponse) CallResponse {
	return CallResponse{
		CallResponse: response,
	}
}

func (r CallResponse) WithMetadata(metadata AppMetadataForClient) CallResponse {
	r.AppMetadata = metadata
	return r
}

func (p *Proxy) Call(in Incoming, creq apps.CallRequest) CallResponse {
	if creq.Context.AppID == "" {
		return NewProxyCallResponse(apps.NewErrorResponse(
			utils.NewInvalidError("app_id is not set in Context, don't know what app to call")))
	}

	app, err := p.store.App.Get(creq.Context.AppID)
	if err != nil {
		return NewProxyCallResponse(apps.NewErrorResponse(err))
	}

	cresp, _ := p.callApp(in, *app, creq)
	return NewProxyCallResponse(cresp).WithMetadata(AppMetadataForClient{
		BotUserID:   app.BotUserID,
		BotUsername: app.BotUsername,
	})
}

func (p *Proxy) call(in Incoming, app apps.App, call apps.Call, cc *apps.Context, valuePairs ...interface{}) apps.CallResponse {
	values := map[string]interface{}{}
	for len(valuePairs) > 0 {
		if len(valuePairs) == 1 {
			return apps.NewErrorResponse(
				errors.Errorf("mismatched parameter count, no value for %v", valuePairs[0]))
		}
		key, ok := valuePairs[0].(string)
		if !ok {
			return apps.NewErrorResponse(
				errors.Errorf("mismatched type %T for key %v, expected string", valuePairs[0], valuePairs[0]))
		}
		values[key] = valuePairs[1]
		valuePairs = valuePairs[2:]
	}

	if cc == nil {
		cc = &apps.Context{}
	}
	cresp, _ := p.callApp(in, app, apps.CallRequest{
		Call:    call,
		Context: *cc,
		Values:  values,
	})
	return cresp
}

func (p *Proxy) callApp(in Incoming, app apps.App, creq apps.CallRequest) (apps.CallResponse, error) {
	respondErr := func(err error) (apps.CallResponse, error) {
		return apps.NewErrorResponse(err), err
	}

	conf, _, log := p.conf.Basic()
	log = log.With("app_id", app.AppID)

	if !p.appIsEnabled(app) {
		return respondErr(errors.Errorf("%s is disabled", app.AppID))
	}

	if creq.Path[0] != '/' {
		return respondErr(utils.NewInvalidError("call path must start with a %q: %q", "/", creq.Path))
	}
	cleanPath, err := utils.CleanPath(creq.Path)
	if err != nil {
		return respondErr(err)
	}
	creq.Path = cleanPath

	up, err := p.upstreamForApp(app)
	if err != nil {
		return respondErr(err)
	}

	cc := creq.Context
	cc = in.updateContext(cc)
	creq.Context, err = p.expandContext(in, app, &cc, creq.Expand)
	if err != nil {
		return respondErr(err)
	}

	cresp, err := upstream.Call(up, app, creq)
	if err != nil {
		return cresp, err
	}
	if cresp.Type == "" {
		cresp.Type = apps.CallResponseTypeOK
	}

	if cresp.Form != nil {
		if cresp.Form.Icon != "" {
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

	return cresp, nil
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
		err := appservices.CheckSubscriptionPermission(&p.conf.MattermostAPI().User, sub)
		if err != nil {
			// Don't log the error it can be to spammy
			continue
		}

		err = p.notifyForSubscription(&base, sub)
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
	if !p.appIsEnabled(*app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}

	creq.Context, err = p.expandContext(Incoming{}, *app, base, sub.Call.Expand)
	if err != nil {
		return err
	}
	creq.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(*app)
	if err != nil {
		return err
	}
	return upstream.Notify(up, *app, creq)
}

func (p *Proxy) NotifyRemoteWebhook(app apps.App, data []byte, webhookPath string) error {
	if !p.appIsEnabled(app) {
		return errors.Errorf("%s is disabled", app.AppID)
	}
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

	conf := p.conf.Get()
	cc := contextForApp(app, apps.Context{}, conf)
	// Set acting user to bot.
	cc.ActingUserID = app.BotUserID
	cc.ActingUserAccessToken = app.BotAccessToken

	// TODO: do we need to customize the Expand & State for the webhook Call?
	return upstream.Notify(up, app, apps.CallRequest{
		Call: apps.Call{
			Path: path.Join(appspath.Webhook, webhookPath),
		},
		Context: cc,
		Values: map[string]interface{}{
			"data": datav,
		},
	})
}

var atMentionRegexp = regexp.MustCompile(`\B@[[:alnum:]][[:alnum:]\.\-_:]*`)

func (p *Proxy) NotifyMessageHasBeenPosted(post *model.Post, cc apps.Context) error {
	postSubs, err := p.store.Subscription.Get(apps.SubjectPostCreated, cc.TeamID, cc.ChannelID)
	if err != nil && err != utils.ErrNotFound {
		return errors.Wrap(err, "failed to get post_created subscriptions")
	}

	subs := []apps.Subscription{}
	subs = append(subs, postSubs...)

	mentions := possibleAtMentions(post.Message)

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
					subs = append(subs, sub)
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

	return p.getStatic(*app, path)
}

func (p *Proxy) getStatic(app apps.App, path string) (io.ReadCloser, int, error) {
	up, err := p.upstreamForApp(app)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return up.GetStatic(app, path)
}

// possibleAtMentions is copied over from mattermost-server/app.possibleAtMentions
func possibleAtMentions(message string) []string {
	var names []string

	if !strings.Contains(message, "@") {
		return names
	}

	alreadyMentioned := make(map[string]bool)
	for _, match := range atMentionRegexp.FindAllString(message, -1) {
		name := model.NormalizeUsername(match[1:])
		if !alreadyMentioned[name] && model.IsValidUsernameAllowRemote(name) {
			names = append(names, name)
			alreadyMentioned[name] = true
		}
	}

	return names
}
