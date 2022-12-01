package main

import (
	"fmt"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var numPostMenuBindings int64 = -1
var numChannelHeaderBindings int64 = -1

func initNumBindingsCommand(r *mux.Router) {
	handleCall(r, NumBindingsPath, handleNumBindingsCommand)
}

func handleNumBindingsCommand(creq *apps.CallRequest) apps.CallResponse {
	location := creq.GetValue("location", "post_menu")
	numStr := creq.GetValue("number", "-1")

	num, _ := strconv.ParseInt(numStr, 10, 32)

	switch location {
	case "post_menu":
		numPostMenuBindings = num
	case "channel_header":
		numChannelHeaderBindings = num
	default:
		return apps.NewErrorResponse(errors.Errorf("unsupported location %s", location))
	}

	return apps.NewTextResponse(fmt.Sprintf("changed %s to %d", location, num))
}
