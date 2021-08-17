// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"encoding/json"
	"path"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v5/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

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
