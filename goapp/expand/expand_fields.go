package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

var userActionExpandFields = []apps.Field{
	expandField("app"),
	expandField("acting_user"),
	expandField("acting_user_access_token"),
	expandField("locale"),
	expandField("channel"),
	expandField("channel_member"),
	expandField("team"),
	expandField("team_member"),
	expandField("post"),
	expandField("root_post"),
	expandField("oauth2_app"),
	expandField("oauth2_user"),
	expandField("user"),
}

func expandField(name string) apps.Field {
	return apps.Field{
		Type: apps.FieldTypeStaticSelect,
		Name: name,
		SelectStaticOptions: []apps.SelectOption{
			{
				Label: string(apps.ExpandNone),
				Value: string(apps.ExpandNone),
			},
			{
				Label: string(apps.ExpandID),
				Value: string(apps.ExpandID),
			},
			{
				Label: string(apps.ExpandSummary),
				Value: string(apps.ExpandSummary),
			},
			{
				Label: string(apps.ExpandAll),
				Value: string(apps.ExpandAll),
			},
		},
	}
}

func expandFromValues(creq goapp.CallRequest) apps.Expand {
	return apps.Expand{
		App:                   apps.ExpandLevel(creq.GetValue("app", "")),
		ActingUser:            apps.ExpandLevel(creq.GetValue("acting_user", "")),
		ActingUserAccessToken: apps.ExpandLevel(creq.GetValue("acting_user_access_token", "")),
		Locale:                apps.ExpandLevel(creq.GetValue("locale", "")),
		Channel:               apps.ExpandLevel(creq.GetValue("channel", "")),
		ChannelMember:         apps.ExpandLevel(creq.GetValue("channel_member", "")),
		Team:                  apps.ExpandLevel(creq.GetValue("team", "")),
		TeamMember:            apps.ExpandLevel(creq.GetValue("team_member", "")),
		Post:                  apps.ExpandLevel(creq.GetValue("post", "")),
		RootPost:              apps.ExpandLevel(creq.GetValue("root_post", "")),
		OAuth2App:             apps.ExpandLevel(creq.GetValue("oauth2_app", "")),
		OAuth2User:            apps.ExpandLevel(creq.GetValue("oauth2_user", "")),
		User:                  apps.ExpandLevel(creq.GetValue("user", "")),
	}
}
