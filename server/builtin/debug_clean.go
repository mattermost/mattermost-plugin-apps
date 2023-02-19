// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) debugCleanCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "clean",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.clean.label",
			Other: "clean",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.clean.description",
			Other: "Remove all Apps and reset the persistent store",
		}),
		Submit: newUserCall(PathDebugClean),
	}
}

func (a *builtinApp) debugClean(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	loc := a.newLocalizer(creq)
	err := a.api.Mattermost.KV.DeleteAll()
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	done := "- " + a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.debug.clean.submit.kv",
		Other: "Deleted all KV records.",
	}) + "\n"

	err = r.Config.StoreConfig(config.StoredConfig{}, r.Log)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	done += "- " + a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
		ID:    "command.debug.clean.submit.config",
		Other: "Emptied the config.",
	}) + "\n"

	return apps.NewTextResponse(done)
}
