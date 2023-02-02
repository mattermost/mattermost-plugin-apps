package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var expandAll = apps.Expand{
	ActingUser:            apps.ExpandAll,
	ActingUserAccessToken: apps.ExpandAll,
	App:                   apps.ExpandAll,
	Channel:               apps.ExpandAll,
	ChannelMember:         apps.ExpandAll,
	Locale:                apps.ExpandAll,
	OAuth2App:             apps.ExpandAll,
	OAuth2User:            apps.ExpandAll,
	Post:                  apps.ExpandAll,
	RootPost:              apps.ExpandAll,
	Team:                  apps.ExpandAll,
	TeamMember:            apps.ExpandAll,
	User:                  apps.ExpandAll,
}

var allSubjects = []apps.Subject{
	// apps.SubjectPostCreated,
	// apps.SubjectSelfMentioned,
	apps.SubjectBotJoinedChannel,
	apps.SubjectBotJoinedTeam,
	apps.SubjectBotLeftChannel,
	apps.SubjectBotLeftTeam,
	apps.SubjectChannelCreated,
	apps.SubjectUserCreated,
	apps.SubjectUserJoinedChannel,
	apps.SubjectUserJoinedTeam,
	apps.SubjectUserLeftChannel,
	apps.SubjectUserLeftTeam,
}

func subscriptionOptions() []apps.SelectOption {
	opts := []apps.SelectOption{}
	for _, subject := range allSubjects {
		opts = append(opts, apps.SelectOption{
			Label: string(subject),
			Value: string(subject),
		})
	}

	return opts
}

func subscriptionCommandBinding(label, callPath string) apps.Binding {
	return apps.Binding{
		Label: label,
		Form: &apps.Form{
			Submit: apps.NewCall(callPath).WithExpand(apps.Expand{
				ActingUserAccessToken: apps.ExpandAll,
				Channel:               apps.ExpandSummary,
				Team:                  apps.ExpandSummary,
			}),
			Fields: []apps.Field{
				{
					Name:                 "subject",
					IsRequired:           true,
					AutocompletePosition: 1,
					Type:                 apps.FieldTypeStaticSelect,
					SelectStaticOptions:  subscriptionOptions(),
				},
				{
					Name: "channel",
					Type: apps.FieldTypeChannel,
				},
				{
					Name: "team_name",
					Type: apps.FieldTypeText,
				},
				{
					Name: "as_bot",
					Type: apps.FieldTypeBool,
				},
			},
		},
	}
}

func initHTTPSubscriptions(r *mux.Router) {
	handleCall(r, NotifyPath, handleNotify)
	handleCall(r, Subscribe, handleSubscribe)
	handleCall(r, Unsubscribe, handleUnsubscribe)
}

func handleSubscription(creq *apps.CallRequest, subscribe bool) apps.CallResponse {
	subject := apps.Subject(creq.GetValue("subject", ""))
	teamName := creq.GetValue("team_name", "")
	channelID := creq.GetValue("channel", "")
	subscribeAsBot := creq.BoolValue("as_bot")
	asActingUser := appclient.AsActingUser(creq.Context)
	subscribeClient := asActingUser
	if subscribeAsBot {
		subscribeClient = appclient.AsBot(creq.Context)
	}

	teamID := ""
	if teamName != "" {
		team, _, err := asActingUser.GetTeamByName(teamName, "")
		if err != nil || team == nil {
			return apps.NewErrorResponse(err)
		}
		teamID = team.Id
	}

	sub := &apps.Subscription{
		Event: apps.Event{
			Subject: subject,
		},
		Call: *apps.NewCall(NotifyPath).WithExpand(expandAll),
	}

	switch subject {
	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel:
		sub.ChannelID = channelID

	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam,
		apps.SubjectChannelCreated:
		sub.TeamID = teamID
		if teamID != "" {
			_, _, err := asActingUser.AddTeamMember(teamID, creq.Context.BotUserID)
			if err != nil {
				return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to team"))
			}
		}
	}

	if !subscribe {
		err := subscribeClient.Unsubscribe(sub)
		if err != nil {
			return apps.NewErrorResponse(err)
		}

		return apps.NewTextResponse("Successfully unsubscribed from `%v` notifications.", subject)
	}

	if creq.Context.Channel.Type == model.ChannelTypeOpen || creq.Context.Channel.Type == model.ChannelTypePrivate {
		_, _, err := asActingUser.AddChannelMember(creq.Context.Channel.Id, creq.Context.BotUserID)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to channel"))
		}
	}

	err := ensureNotifyChannel(creq)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	err = subscribeClient.Subscribe(sub)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	return apps.NewTextResponse("Successfully subscribed to `%v` notifications.", subject)
}

func handleSubscribe(creq *apps.CallRequest) apps.CallResponse {
	return handleSubscription(creq, true)
}

func handleUnsubscribe(creq *apps.CallRequest) apps.CallResponse {
	return handleSubscription(creq, false)
}

func ensureNotifyChannel(creq *apps.CallRequest) error {
	client := appclient.AsActingUser(creq.Context)

	channel, _, err := client.GetChannelByName("test-app-notifications", creq.Context.Team.Id, "")
	if err != nil {
		appErr, ok := err.(*model.AppError)
		if !ok || appErr.StatusCode != http.StatusNotFound {
			return errors.Wrap(err, "failed to look up notification channel")
		}
	}

	if channel == nil {
		channel, _, err = client.CreateChannel(&model.Channel{
			TeamId:      creq.Context.Team.Id,
			Type:        model.ChannelTypePrivate,
			DisplayName: "Test App Notifications",
			Name:        "test-app-notifications",
		})
		if err != nil {
			return errors.Wrap(err, "failed to create up notification channel")
		}
	}

	_, _, err = client.AddChannelMember(channel.Id, creq.Context.BotUserID)
	if err != nil {
		return errors.Wrap(err, "failed to add bot to notification channel")
	}

	return nil
}

const testTeamName = "ad-1"
const testChannelName = "test-app-notifications"

func handleNotify(creq *apps.CallRequest) apps.CallResponse {
	client := appclient.AsBot(creq.Context)

	team, _, err := client.GetTeamByName(testTeamName, "")
	if err != nil {
		Log.Debugf("failed to look up team %s", testTeamName, err)
	}

	channel, _, err := client.GetChannelByName(testChannelName, team.Id, "")
	if err != nil {
		Log.Debugf("failed to look up notification channel: %v", err)
	}

	post := &model.Post{
		Message: fmt.Sprintf("received notification:\n```\n%s\n```\n", utils.Pretty(creq.Context)),
	}
	// Post the notification to the global notification channel
	if channel != nil {
		post.ChannelId = channel.Id

		_, err = client.CreatePost(post)
		if err != nil {
			Log.Debugf("failed to create post in global channel: %v", err)
		}
	}

	// CC the notification to the relevant channel if possible.
	if creq.Context.Channel != nil {
		post.ChannelId = creq.Context.Channel.Id
		if creq.Context.Post != nil {
			post.RootId = creq.Context.Post.Id
			if creq.Context.Post.RootId != "" {
				post.RootId = creq.Context.Post.RootId
			}
		}

		_, err = client.CreatePost(post)
		if err != nil {
			Log.Debugf("failed to create post in channel: %v", err)
		}
	}

	return apps.NewTextResponse("OK")
}
