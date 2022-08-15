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

var allSubjects = map[apps.Subject]apps.Expand{
	apps.SubjectUserCreated: {
		User: apps.ExpandAll,
		Team: apps.ExpandAll,
	},
	apps.SubjectBotJoinedChannel: {
		User:          apps.ExpandAll,
		Channel:       apps.ExpandAll,
		ChannelMember: apps.ExpandAll,
	},
	apps.SubjectBotLeftChannel: {
		User:          apps.ExpandAll,
		Channel:       apps.ExpandAll,
		ChannelMember: apps.ExpandAll,
	},
	apps.SubjectBotJoinedTeam: {
		User:       apps.ExpandAll,
		Team:       apps.ExpandAll,
		TeamMember: apps.ExpandAll,
	},
	apps.SubjectBotLeftTeam: {
		User:       apps.ExpandAll,
		Team:       apps.ExpandAll,
		TeamMember: apps.ExpandAll,
	},
	// apps.SubjectBotMentioned: {
	// 	Channel:  apps.ExpandAll,
	// 	Post:     apps.ExpandSummary,
	// 	RootPost: apps.ExpandSummary,
	// },
	apps.SubjectUserJoinedChannel: {
		User:          apps.ExpandAll,
		Channel:       apps.ExpandAll,
		ChannelMember: apps.ExpandAll,
	},
	apps.SubjectUserLeftChannel: {
		User:          apps.ExpandAll,
		Channel:       apps.ExpandAll,
		ChannelMember: apps.ExpandAll,
	},
	// apps.SubjectPostCreated: {
	// 	Post:    apps.ExpandAll,
	// 	Channel: apps.ExpandAll,
	// },
	apps.SubjectUserJoinedTeam: {
		User:       apps.ExpandAll,
		Team:       apps.ExpandAll,
		TeamMember: apps.ExpandAll,
	},
	apps.SubjectUserLeftTeam: {
		User:       apps.ExpandAll,
		Team:       apps.ExpandAll,
		TeamMember: apps.ExpandAll,
	},
	apps.SubjectChannelCreated: {
		Channel: apps.ExpandAll,
	},
}

func subscriptionOptions() []apps.SelectOption {
	opts := []apps.SelectOption{}
	for subject := range allSubjects {
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
					Label:                "subject",
					IsRequired:           true,
					AutocompletePosition: 1,
					Type:                 apps.FieldTypeStaticSelect,
					SelectStaticOptions:  subscriptionOptions(),
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
	client := appclient.AsActingUser(creq.Context)

	sub := &apps.Subscription{
		Event: apps.Event{
			Subject: subject,
		},
		Call: *apps.NewCall(NotifyPath).WithExpand(allSubjects[subject]),
	}

	switch subject {
	case apps.SubjectUserJoinedChannel,
		apps.SubjectUserLeftChannel:
		sub.ChannelID = creq.Context.Channel.Id

	case apps.SubjectUserJoinedTeam,
		apps.SubjectUserLeftTeam,
		apps.SubjectBotJoinedChannel,
		apps.SubjectBotLeftChannel,
		apps.SubjectChannelCreated:
		sub.TeamID = creq.Context.Team.Id
	}

	if !subscribe {
		err := client.Unsubscribe(sub)
		if err != nil {
			return apps.NewErrorResponse(err)
		}

		return apps.NewTextResponse("Successfully unsubscribed from `%v` notifications.", subject)
	}

	if creq.Context.Team != nil {
		_, _, err := client.AddTeamMember(creq.Context.Team.Id, creq.Context.BotUserID)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to team"))
		}
	}

	if creq.Context.Channel.Type == model.ChannelTypeOpen || creq.Context.Channel.Type == model.ChannelTypePrivate {
		_, _, err := client.AddChannelMember(creq.Context.Channel.Id, creq.Context.BotUserID)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to channel"))
		}
	}

	err := ensureNotifyChannel(creq)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	err = client.Subscribe(sub)
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

func handleNotify(creq *apps.CallRequest) apps.CallResponse {
	client := appclient.AsBot(creq.Context)

	teamID := ""

	switch {
	case creq.Context.Team != nil:
		teamID = creq.Context.Team.Id

	case creq.Context.Channel != nil:
		teamID = creq.Context.Channel.TeamId
	}

	channel, _, err := client.GetChannelByName("test-app-notifications", teamID, "")
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
