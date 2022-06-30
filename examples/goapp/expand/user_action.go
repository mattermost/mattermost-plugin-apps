package main

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

//userAction defines /user-action command, the app's only channel header button,
//and a Post Menu item. The command can be autocompleted, or filled as a form;
//the other locations pop a modal dialog to customize the subsequent Expand.
func userAction() goapp.Bindable {
	return goapp.MakeBindableFormOrPanic(
		"user-action",
		apps.Form{
			Title:  "Test how Expand works on user actions",
			Header: "TODO",
			Fields: userActionExpandFields,
		},
		func(creq goapp.CallRequest) apps.CallResponse {
			expand := expandFromValues(creq)
			expand.ActingUserAccessToken = apps.ExpandAll
			submit := apps.NewCall("/echo").WithExpand(expand)

			return apps.NewFormResponse(apps.Form{
				Title: "Example of a user call with expand",
				Header: fmt.Sprintf("Press OK to submit the following call: %s\n\n", utils.JSONBlock(submit)) +
					fmt.Sprintf("Note that `\"acting_user_access_token\":\"all\"` is added so that team and channel expansion works without first having to add the bot as the team/channel member."),
				Submit: submit,
			})
		},
	)
}
