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
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/upstream"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (p *Proxy) Notify(base apps.Context, subj apps.Subject) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	mm := p.conf.MattermostAPI()
	r := incoming.NewRequest(mm, p.conf, utils.NewPluginLogger(mm), p.sessionService, incoming.WithCtx(ctx))

	subs, err := p.store.Subscription.Get(subj, base.TeamID, base.ChannelID)
	if err != nil {
		return err
	}

	return p.notify(r, base, subs)
}

func (p *Proxy) notify(r *incoming.Request, base apps.Context, subs []apps.Subscription) error {
	for _, sub := range subs {
		// TODO(Ben): Convert to use App session
		err := appservices.CheckSubscriptionPermission(&p.conf.MattermostAPI().User, sub, base.ChannelID, base.TeamID)
		if err != nil {
			// Don't log the error it can be to spammy
			continue
		}

		r = r.Clone()
		r.SetAppID(sub.AppID)
		r.Log = r.Log.With("subject", sub.Subject)

		err = p.notifyForSubscription(r, &base, sub)
		if err != nil {
			r.Log.WithError(err).Debugf("failed to send subscription notification to app")
		}
	}

	return nil
}

func (p *Proxy) notifyForSubscription(r *incoming.Request, base *apps.Context, sub apps.Subscription) error {
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

	expanded, err := p.expandContext(r, *app, base, sub.Call.Expand)
	if err != nil {
		return err
	}
	creq.Context = *expanded
	creq.Context.Subject = sub.Subject

	up, err := p.upstreamForApp(*app)
	if err != nil {
		return err
	}
	return upstream.Notify(r.Ctx(), up, *app, creq)
}

func (p *Proxy) NotifyMessageHasBeenPosted(post *model.Post, cc apps.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	mm := p.conf.MattermostAPI()
	r := incoming.NewRequest(mm, p.conf, utils.NewPluginLogger(mm), p.sessionService, incoming.WithCtx(ctx))

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

	return p.notify(r, cc, subs)
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
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	mm := p.conf.MattermostAPI()
	r := incoming.NewRequest(mm, p.conf, utils.NewPluginLogger(mm), p.sessionService, incoming.WithCtx(ctx))

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

	return p.notify(r, cc, subs)
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
