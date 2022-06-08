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

func (p *Proxy) NotifyUserJoinedChannel(channelID, userID string) {
	p.notifyJoinLeave("", channelID, userID, apps.SubjectUserJoinedChannel, apps.SubjectBotJoinedChannel)
}

func (p *Proxy) NotifyUserLeftChannel(channelID, userID string) {
	p.notifyJoinLeave("", channelID, userID, apps.SubjectUserLeftChannel, apps.SubjectBotLeftChannel)
}

func (p *Proxy) NotifyUserJoinedTeam(teamID, userID string) {
	p.notifyJoinLeave(teamID, "", userID, apps.SubjectUserJoinedTeam, apps.SubjectBotJoinedTeam)
}

func (p *Proxy) NotifyUserLeftTeam(teamID, userID string) {
	p.notifyJoinLeave(teamID, "", userID, apps.SubjectUserLeftTeam, apps.SubjectBotLeftTeam)
}

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
	err = p.callApp(appRequest, app, creq, true)
}

func (p *Proxy) notifyJoinLeave(teamID, channelID, userID string, subject, botSubject apps.Subject) {
	user, err := p.conf.MattermostAPI().User.Get(userID)
	if err != nil {
		p.log.WithError(err).Errorf("Notify join/leave: failed to get user")
		return
	}

	event := apps.Event{
		Subject:   subject,
		ChannelID: channelID,
		TeamID:    teamID,
	}
	uac := apps.UserAgentContext{
		UserID:    user.Id,
		TeamID:    teamID,
		ChannelID: channelID,
	}

	p.notify(nil, event, uac)

	// If the user is a bot, process SubjectBotLeftChannel; only notify the app
	// with the matching BotUserID.
	if user.IsBot {
		allApps := p.store.App.AsMap()
		event.Subject = botSubject
		p.notify(func(sub store.Subscription) bool {
			if app, ok := allApps[sub.AppID]; ok {
				return app.BotUserID == userID
			}
			return false
		}, event, uac)
	}
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
