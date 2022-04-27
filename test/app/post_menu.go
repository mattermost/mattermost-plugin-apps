package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func postMenuBindings(_ apps.Context) []apps.Binding {
	out := []apps.Binding{}
	out = append(out, validResponseBindings...)
	out = append(out, errorResponseBindings...)
	out = append(out, validInputBindings...)

	if IncludeInvalid {
		out = append(out, invalidResponseBindings...)
		out = append(out, invalidFormBindings...)
	}

	return out
}
