package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/appclient"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func timerCommandBinding(label, callPath string) apps.Binding {
	return apps.Binding{
		Label: label,
		Form: &apps.Form{
			Submit: apps.NewCall(callPath).WithExpand(apps.Expand{
				ActingUserAccessToken: apps.ExpandAll,
				Channel:               apps.ExpandID,
				Team:                  apps.ExpandID,
			}),
			Fields: []apps.Field{
				{
					Name:                 "duration",
					Label:                "duration",
					Description:          "duration until the timer expires in seconds",
					IsRequired:           true,
					AutocompletePosition: 1,
					Type:                 apps.FieldTypeText,
					TextSubtype:          apps.TextFieldSubtypeNumber,
				}, {
					Name:                 "state",
					Label:                "state",
					Description:          "a state to return",
					IsRequired:           false,
					AutocompletePosition: 2,
					Type:                 apps.FieldTypeText,
					TextSubtype:          apps.TextFieldSubtypeInput,
				},
			},
		},
	}
}

func initHTTPTimer(r *mux.Router) {
	handleCall(r, ExecuteTimer, handleExecuteTimer)
	handleCall(r, CreateTimer, handleCreateTimer)
}

func handleCreateTimer(creq *apps.CallRequest) apps.CallResponse {
	client := appclient.AsActingUser(creq.Context)

	durationString := creq.GetValue("duration", "")
	durcation, err := strconv.Atoi(durationString)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "duration is not a number"))
	}
	at := time.Now().Add(time.Second * time.Duration(durcation))

	state := creq.GetValue("state", "")

	t := &apps.Timer{
		At:   at.UnixMilli(),
		Call: *apps.NewCall(ExecuteTimer).WithExpand(apps.Expand{}).WithState(state),
	}
	if creq.Context.Channel != nil {
		t.ChannelID = creq.Context.Channel.Id
		t.Call.Expand.Channel = apps.ExpandAll
	}
	if creq.Context.Team != nil {
		t.TeamID = creq.Context.Team.Id
		t.Call.Expand.Team = apps.ExpandAll
	}

	if creq.Context.Team != nil {
		_, _, err = client.AddTeamMember(creq.Context.Team.Id, creq.Context.BotUserID)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to team"))
		}
	}

	if creq.Context.Channel.Type == model.ChannelTypeOpen || creq.Context.Channel.Type == model.ChannelTypePrivate {
		_, _, err = client.AddChannelMember(creq.Context.Channel.Id, creq.Context.BotUserID)
		if err != nil {
			return apps.NewErrorResponse(errors.Wrap(err, "failed to add bot to channel"))
		}
	}

	err = client.CreateTimer(t)
	if err != nil {
		return apps.NewErrorResponse(errors.Wrap(err, "failed to create timer"))
	}

	return apps.NewTextResponse("Successfully set a timer to `%v`.", at.String())
}

func handleExecuteTimer(creq *apps.CallRequest) apps.CallResponse {
	client := appclient.AsBot(creq.Context)

	post := &model.Post{
		Message: fmt.Sprintf("Received timer,\n`State`: `%s`,\n`Context`: \n```json\n%s\n```\n", creq.State, utils.Pretty(creq.Context)),
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

		_, err := client.CreatePost(post)
		if err != nil {
			Log.Debugf("failed to create post in channel: %v", err)
		}
	}

	return apps.NewTextResponse("OK")
}
