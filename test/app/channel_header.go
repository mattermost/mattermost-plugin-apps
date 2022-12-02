package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func channelHeaderBindings(_ apps.Context) []apps.Binding {
	out := []apps.Binding{}
	out = append(out, validResponseBindings...)
	out = append(out, errorResponseBindings...)
	out = append(out, validInputBindings...)

	if IncludeInvalid {
		out = append(out, invalidResponseBindings...)
		out = append(out, invalidFormBindings...)
	}

	if numChannelHeaderBindings < 0 {
		return out
	}

	return out[0:numChannelHeaderBindings]
}
