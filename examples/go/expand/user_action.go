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
var userAction = goapp.MakeBindableFormOrPanic("user-action",
	handleUserAction,
	apps.Form{
		Title:  "Test how Expand works on user actions",
		Header: "TODO",
		Fields: expandFields,
	})

// handleUserAction processes the submit
func handleUserAction(creq goapp.CallRequest) apps.CallResponse {
	submit := apps.NewCall("/echo").
		WithExpand(expandFromValues(creq))

	return apps.NewFormResponse(apps.Form{
		Title:  "Example of a user call with expand",
		Header: fmt.Sprintf("Press OK to submit the following call: %s", utils.JSONBlock(submit)),
		Submit: submit,
	})
}
