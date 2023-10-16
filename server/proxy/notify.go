// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"

	"github.com/mattermost/mattermost/server/public/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

// NotifyUserCreated handles plugin's UserHasBeenCreated callback. It emits
// "user_created" notifications to subscribed apps.
func (p *Proxy) NotifyUserCreated(userID string) {
	p.notifyAll(
		apps.Event{
			Subject: apps.SubjectUserCreated,
		},
		apps.UserAgentContext{
			UserID: userID,
		},
		nil, // no special filtering, notify all subscriptions mathcing the event.
		nil, // no special expand rules.
	)
}

func (p *Proxy) NotifyUserChannel(member *model.ChannelMember, actor *model.User, joined bool) {
	subject := apps.SubjectUserJoinedChannel
	if !joined {
		subject = apps.SubjectUserLeftChannel
	}

	log := p.conf.NewBaseLogger().With("subject", subject)

	mm := p.conf.MattermostAPI()
	user, err := mm.User.Get(member.UserId)
	if err != nil {
		log.WithError(err).Debugf("failed to get user")
		return
	}
	channel, err := mm.Channel.Get(member.ChannelId)
	if err != nil {
		log.WithError(err).Debugf("%s: failed to get channel", subject)
		return
	}

	// Notify on user_joined|left_channel subscriptions specific to the channel
	// that may include any user.
	p.notifyAll(
		apps.Event{
			Subject:   subject,
			ChannelID: channel.Id,
		},
		apps.UserAgentContext{
			ChannelID: channel.Id,
			TeamID:    channel.TeamId,
			UserID:    user.Id,
		},
		nil, // no special filtering, notify all subscriptions matching the event.
		nil, // no special expand rules for "any user" subscriptions.
	)

	// Notify on "self" subscriptions for the user.
	p.notifyAll(
		apps.Event{
			Subject: subject,
		},
		apps.UserAgentContext{
			ChannelID: channel.Id,
			TeamID:    channel.TeamId,
			UserID:    member.UserId,
		},
		// Include subscriptions where the user (added/removed channel member)
		// is the same as the subscriber.
		func(sub store.Subscription) bool {
			return sub.OwnerUserID == member.UserId
		},
		// special expand for "self" subscriptions.
		newExpandSelfGetter(mm, user, member, nil, channel),
	)

	// Notify on the deprecated bot_joined|left_channel subscriptions.
	if user.IsBot {
		allApps := p.store.App.AsMap(store.EnabledAppsOnly)
		subject = apps.SubjectBotJoinedChannel
		if !joined {
			subject = apps.SubjectBotLeftChannel
		}
		p.notifyAll(
			apps.Event{
				Subject: subject,
			},
			apps.UserAgentContext{
				ChannelID: channel.Id,
				TeamID:    channel.TeamId,
				UserID:    member.UserId,
			},
			func(sub store.Subscription) bool {
				if app, ok := allApps[sub.AppID]; ok {
					return app.BotUserID == member.UserId
				}
				return false
			},
			// special expand for "self" subscriptions.
			newExpandSelfGetter(mm, user, member, nil, channel),
		)
	}
}

func (p *Proxy) NotifyUserTeam(member *model.TeamMember, actor *model.User, joined bool) {
	subject := apps.SubjectUserJoinedTeam
	if !joined {
		subject = apps.SubjectUserLeftTeam
	}
	log := p.conf.NewBaseLogger().With("subject", subject)

	mm := p.conf.MattermostAPI()
	user, err := mm.User.Get(member.UserId)
	if err != nil {
		log.WithError(err).Debugf("%s: failed to get user %s", subject, member.UserId)
		return
	}

	// TODO: add a test: user subscribed to user_joined_team, then removed from the
	// team. Subsequent notifications should fail.

	// TODO: OnUserLeftTeam can remove any team-specific subscriptions for the
	// user.

	// Notify on user_joined|left_team subscriptions specific to the team that
	// may include any user.
	p.notifyAll(
		apps.Event{
			Subject: subject,
			TeamID:  member.TeamId,
		},
		apps.UserAgentContext{
			UserID: user.Id,
			TeamID: member.TeamId,
		},
		nil, // no special filtering, notify all subscriptions matching the event.
		nil, // no special expand rules for "any user" subscriptions.
	)

	p.notifyAll(
		apps.Event{
			Subject: subject,
		},
		apps.UserAgentContext{
			TeamID: member.TeamId,
			UserID: member.UserId,
		},
		// Include subscriptions where the user (added/removed team member)
		// is the same as the subscriber.
		func(sub store.Subscription) bool {
			return sub.OwnerUserID == member.UserId
		},
		// special expand for "self" subscriptions.
		newExpandSelfGetter(mm, user, nil, member, nil),
	)

	// If the user is a bot, process SubjectBot...Channel; only notify the app
	// with the matching BotUserID.
	if user.IsBot {
		allApps := p.store.App.AsMap(store.EnabledAppsOnly)
		subject = apps.SubjectBotJoinedTeam
		if !joined {
			subject = apps.SubjectBotLeftTeam
		}
		p.notifyAll(
			apps.Event{
				Subject: subject,
			},
			apps.UserAgentContext{
				UserID: user.Id,
				TeamID: member.TeamId,
			},
			func(sub store.Subscription) bool {
				if app, ok := allApps[sub.AppID]; ok {
					return app.BotUserID == member.UserId
				}
				return false
			},
			// special expand for "self" subscriptions.
			newExpandSelfGetter(mm, user, nil, member, nil),
		)
	}
}

// NotifyChannelCreated handles plugin's ChannelHasBeenCreated callback. It emits
// "channel_created" notifications to subscribed apps.
func (p *Proxy) NotifyChannelCreated(teamID, channelID string) {
	// If the newly created channel is a DM, there is no teamID.
	// Do not notify apps in this case.
	if teamID == "" {
		return
	}

	p.notifyAll(
		apps.Event{
			Subject: apps.SubjectChannelCreated,
			TeamID:  teamID,
		},
		apps.UserAgentContext{
			TeamID:    teamID,
			ChannelID: channelID,
		},
		nil, // no special filtering, notify all subscriptions matching the event.
		nil, // no special expand logic.
	)
}

func (p *Proxy) notifyAll(event apps.Event, uac apps.UserAgentContext, match func(store.Subscription) bool, getter ExpandGetter) {
	log := p.conf.NewBaseLogger().With("event", event)

	subs, err := p.store.Subscription.Get(event)
	if err != nil {
		log.WithError(err).Errorf("notify: failed to load subscriptions")
		return
	}

	for _, sub := range subs {
		if match == nil || match(sub) {
			sub := sub
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
				defer cancel()
				r := p.NewIncomingRequest().WithCtx(ctx)
				r.Log = log

				p.invokeNotify(r, event, sub,
					&apps.Context{
						Subject:          event.Subject,
						UserAgentContext: uac,
					},
					getter)
			}()
		}
	}
}

func (p *Proxy) invokeNotify(r *incoming.Request, event apps.Event, sub store.Subscription, contextToExpand *apps.Context, getter ExpandGetter) {
	var err error
	defer func() {
		if err == nil {
			return
		}
		if p.conf.Get().DeveloperMode {
			r.Log.WithError(err).Errorf("notify error")
		} else {
			r.Log.Debugf("notify")
		}
	}()

	app, err := p.GetInstalledApp(sub.AppID, true)
	if err != nil {
		return
	}

	if contextToExpand == nil {
		contextToExpand = &apps.Context{
			UserAgentContext: apps.UserAgentContext{
				TeamID:    event.TeamID,
				ChannelID: event.ChannelID,
			},
		}
	}

	appRequest := r.WithDestination(sub.AppID)
	appRequest = appRequest.WithActingUserID(sub.OwnerUserID)
	creq := apps.CallRequest{
		Call:    sub.Call,
		Context: *contextToExpand,
	}
	r.Log = r.Log.With(creq)
	cresp := p.callAppWithExpandGetter(appRequest, app, creq, true, getter)
	if cresp.Type == apps.CallResponseTypeError {
		err = cresp
	}
	r.Log = r.Log.With(cresp)
}

// func (p *Proxy) NotifyMessageHasBeenPosted(post *model.Post, cc apps.Context) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
// 	defer cancel()

// 	mm := p.conf.MattermostAPI()
// 	r := incoming.NewRequest(p.conf, utils.NewPluginLogger(mm), p.sessionService).WithCtx(ctx)

// 	subs, err := p.store.Subscription.Get(apps.SubjectPostCreated, cc.TeamID, cc.ChannelID)
// 	if err != nil && err != utils.ErrNotFound {
// 		return errors.Wrap(err, "failed to get post_created subscriptions")
// 	}

// 	appMentions := []string{}
// 	allApps := p.store.App.AsMap()
// 	for _, m := range possibleAtMentions(post.Message) {
// 		for _, app := range allApps {
// 			if app.BotUsername == m {
// 				appMentions = append(appMentions, m)
// 			}
// 		}
// 	}
// 	if len(appMentions) > 0 {
// 		mentionSubs, err := p.store.Subscription.Get(apps.SubjectBotMentioned, cc.TeamID, cc.ChannelID)
// 		if err != nil && err != utils.ErrNotFound {
// 			return errors.Wrap(err, "failed to get bot_mentioned subscriptions")
// 		}
// 		for _, sub := range mentionSubs {
// 			for _, mention := range appMentions {
// 				app, ok := allApps[sub.AppID]
// 				if ok && mention == app.BotUsername {
// 					subs = append(subs, sub)
// 				}
// 			}
// 		}
// 	}

// 	if len(subs) == 0 {
// 		return nil
// 	}

// 	return p.notify(r, cc, subs)
// }

// var atMentionRegexp = regexp.MustCompile(`\B@[[:alnum:]][[:alnum:]\.\-_:]*`)

// // possibleAtMentions is copied over from mattermost-server/app.possibleAtMentions
// func possibleAtMentions(message string) []string {
// 	var names []string

// 	if !strings.Contains(message, "@") {
// 		return names
// 	}

// 	alreadyMentioned := make(map[string]bool)
// 	for _, match := range atMentionRegexp.FindAllString(message, -1) {
// 		name := model.NormalizeUsername(match[1:])
// 		if !alreadyMentioned[name] && model.IsValidUsernameAllowRemote(name) {
// 			names = append(names, name)
// 			alreadyMentioned[name] = true
// 		}
// 	}

// 	return names
// }
