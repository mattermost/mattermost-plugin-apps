// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) debugCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "debug",
		Label:    a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
						ID:    "command.debug.label",
						Other: "debug",
					}),
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
		},
	}
}
