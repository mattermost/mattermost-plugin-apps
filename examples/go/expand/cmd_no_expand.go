package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var noExpand = goapp.Bindable{
	Name:        "no-expand",
	Description: "Has no Expand in its submit",
	BaseSubmit:  apps.NewCall(""),

	Handler: func(creq goapp.CallRequest) apps.CallResponse {
		return apps.NewTextResponse("context:" + utils.PrettyBlock(creq.CallRequest.Context))
	},
}
