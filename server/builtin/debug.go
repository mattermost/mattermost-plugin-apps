// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) debugCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "debug",
		Label:    a.conf.Local(loc, "command.debug.label"),
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
		},
	}
}
