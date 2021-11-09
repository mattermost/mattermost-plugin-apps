// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"
	"regexp"
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/appservices"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/proxy/request"
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
		err := appservices.CheckSubscriptionPermission(&p.conf.MattermostAPI().User, sub, base.ChannelID, base.TeamID)
		if err != nil {
			// Don't log the error it can be to spammy
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
		defer cancel()

		c := request.NewContext(p.conf.MattermostAPI(), p.conf, p.sessionService, request.WithAppID(sub.AppID), request.WithCtx(ctx))
		c.Log = c.Log.With("subject", sub.Subject)

		err = p.notifyForSubscription(c, &base, sub)
		if err != nil {
			c.Log.WithError(err).Debugw("Error sending subscription notification to app")
		}
	}

	return nil
}

func (p *Proxy) notifyForSubscription(c *request.Context, base *apps.Context, sub apps.Subscription) error {
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

	creq.Context, err = p.expandContext(c, *app, base, sub.Call.Expand)
	if err != nil {
		return err
	}
	creq.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(*app)
	if err != nil {
		return err
	}
	return upstream.Notify(c.Ctx(), up, *app, creq)
}

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

var atMentionRegexp = regexp.MustCompile(`\B@[[:alnum:]][[:alnum:]\.\-_:]*`)

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
