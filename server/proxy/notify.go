// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"context"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

// NotifyUserCreated handles plugin's UserHasBeenCreated callback. It emits
// "user_created" notifications to subscribed apps.
func (p *Proxy) NotifyUserCreated(userID string) {
	p.notify(nil,
		apps.Event{
			Subject: apps.SubjectUserCreated,
		},
		apps.UserAgentContext{
			UserID: userID,
		},
	)
}

// NotifyUserJoinedChannel handles plugin's UserHasJoinedChannel callback. It
// emits "user_joined_channel" and "bot_joined_channel" notifications to
// subscribed apps.
func (p *Proxy) NotifyUserJoinedChannel(channelID, userID string) {
	p.notifyUserChannel(channelID, userID, true, "NotifyUserJoinedChannel")
}

// NotifyUserLeftChannel handles plugin's UserHasLeftChannel callback. It emits
// "user_left_channel" and "bot_left_channel" notifications to subscribed apps.
func (p *Proxy) NotifyUserLeftChannel(channelID, userID string) {
	p.notifyUserChannel(channelID, userID, false, "NotifyUserLeftChannel")
}

func (p *Proxy) notifyUserChannel(channelID, userID string, joined bool, method string) {
	mm := p.conf.MattermostAPI()
	log := p.conf.NewBaseLogger().With("method", method)
	user, err := mm.User.Get(userID)
	if err != nil {
		log.WithError(err).Debugf("failed to get user")
		return
	}
	channel, err := mm.Channel.Get(channelID)
	if err != nil {
		log.WithError(err).Debugf("%s: failed to get channel", method)
		return
	}

	subject := apps.SubjectUserJoinedChannel
	if !joined {
		subject = apps.SubjectUserLeftChannel
	}
	p.notify(
		nil,
		apps.Event{
			Subject:   subject,
			ChannelID: channelID,
		},
		apps.UserAgentContext{
			ChannelID: channelID,
			TeamID:    channel.TeamId,
			UserID:    user.Id,
		},
	)
	if !user.IsBot {
		return
	}

	// If the user is a bot, process SubjectBot...Channel; only notify the
	// app with the matching BotUserID.
	allApps := p.store.App.AsMap(store.EnabledAppsOnly)
	subject = apps.SubjectBotJoinedChannel
	if !joined {
		subject = apps.SubjectBotLeftChannel
	}
	p.notify(
		func(sub store.Subscription) bool {
			if app, ok := allApps[sub.AppID]; ok {
				return app.BotUserID == userID
			}
			return false
		},
		apps.Event{
			Subject: subject,
			TeamID:  channel.TeamId,
		},
		apps.UserAgentContext{
			ChannelID: channelID,
			TeamID:    channel.TeamId,
			UserID:    user.Id,
		},
	)
}

// NotifyUserJoinedTeam handles plugin's UserHasJoinedTeam callback. It emits
// "user_joined_team" and "bot_joined_team" notifications to subscribed apps.
func (p *Proxy) NotifyUserJoinedTeam(teamID, userID string) {
	p.notifyUserTeam(teamID, userID, true, "NotifyUserJoinedTeam")
}

// NotifyUserLeftTeam handles plugin's UserHasLeftTeam callback. It emits
// "user_left_team" and "bot_left_team" notifications to subscribed apps.
func (p *Proxy) NotifyUserLeftTeam(teamID, userID string) {
	p.notifyUserTeam(teamID, userID, false, "NotifyUserLeftTeam")
}

func (p *Proxy) notifyUserTeam(teamID, userID string, joined bool, method string) {
	mm := p.conf.MattermostAPI()
	log := p.conf.NewBaseLogger().With("method", method)
	user, err := mm.User.Get(userID)
	if err != nil {
		log.WithError(err).Debugf("%s: failed to get user", method)
		return
	}
	subject := apps.SubjectUserJoinedTeam
	if !joined {
		subject = apps.SubjectUserLeftTeam
	}
	p.notify(
		nil,
		apps.Event{
			Subject: subject,
			TeamID:  teamID,
		},
		apps.UserAgentContext{
			UserID: user.Id,
			TeamID: teamID,
		},
	)
	if !user.IsBot {
		return
	}

	// If the user is a bot, process SubjectBot...Channel; only notify the app
	// with the matching BotUserID.
	allApps := p.store.App.AsMap(store.EnabledAppsOnly)
	subject = apps.SubjectBotJoinedTeam
	if !joined {
		subject = apps.SubjectBotLeftTeam
	}
	p.notify(
		func(sub store.Subscription) bool {
			if app, ok := allApps[sub.AppID]; ok {
				return app.BotUserID == userID
			}
			return false
		},
		apps.Event{
			Subject: subject,
		},
		apps.UserAgentContext{
			UserID: user.Id,
			TeamID: teamID,
		},
	)
}

// NotifyChannelCreated handles plugin's ChannelHasBeenCreated callback. It emits
// "channel_created" notifications to subscribed apps.
func (p *Proxy) NotifyChannelCreated(teamID, channelID string) {
	p.notify(nil,
		apps.Event{
			Subject: apps.SubjectChannelCreated,
			TeamID:  teamID,
		},
		apps.UserAgentContext{
			TeamID:    teamID,
			ChannelID: channelID,
		})
}

func (p *Proxy) notify(match func(store.Subscription) bool, event apps.Event, uac apps.UserAgentContext) {
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()
	r := p.NewIncomingRequest().WithCtx(ctx)
	r.Log = r.Log.With(event)

	subs, err := p.store.Subscription.Get(event)
	if err != nil {
		r.Log.WithError(err).Errorf("notify: failed to load subscriptions")
		return
	}

	for _, sub := range subs {
		if match == nil || match(sub) {
			go p.invokeNotify(r, event, sub, &apps.Context{
				Subject:          event.Subject,
				UserAgentContext: uac,
			})
		}
	}
}

func (p *Proxy) invokeNotify(r *incoming.Request, event apps.Event, sub store.Subscription, contextToExpand *apps.Context) {
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
	cresp := p.callApp(appRequest, app, creq, true)
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
