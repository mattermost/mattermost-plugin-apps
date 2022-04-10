package main

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var callOK = apps.NewCall(OK)

func initHTTPOK(r *mux.Router) {
	handleCall(r, OK, handleOK)
	handleCall(r, OKEmpty, handleOKEmpty)
}

func handleOK(creq *apps.CallRequest) apps.CallResponse {
	return apps.NewTextResponse("```\n%s\n```\n", utils.Pretty(creq))
}

func handleOKEmpty(_ *apps.CallRequest) apps.CallResponse {
	return apps.NewTextResponse("")
}
