// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

const (
	LookupAny              = ""
	LookupDisabledApps     = "disabled"
	LookupEnabledApps      = "enabled"
	LookupInstalledApps    = "installed"
	LookupNotInstalledApps = "not_installed"
)

func (a *builtinApp) appIDField(lookupType string, autocompletePos int, isRequired bool, loc *i18n.Localizer) apps.Field {
	lookupCall := newUserCall(pLookupAppID)
	lookupCall.State = lookupType

	return apps.Field{
		Name: fAppID,
		Type: apps.FieldTypeDynamicSelect,
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.appID.description",
			Other: "Select an App or enter the App ID",
		}),
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.appID.label",
			Other: "app",
		}),
		AutocompletePosition: autocompletePos,
		IsRequired:           isRequired,
		SelectDynamicLookup:  lookupCall,
	}
}

func (a *builtinApp) lookupAppID(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	filter, _ := creq.State.(string)
	var options []apps.SelectOption
	marketplaceApps := a.proxy.GetListedApps(creq.GetValue(fAppID, ""), true)
	for _, app := range marketplaceApps {
		includef := func() bool {
			switch filter {
			case LookupAny:
				return true
			case LookupDisabledApps:
				return app.Installed && !app.Enabled
			case LookupEnabledApps:
				return app.Installed && app.Enabled
			case LookupInstalledApps:
				return app.Installed
			case LookupNotInstalledApps:
				return !app.Installed
			}
			return false
		}

		if includef() {
			options = append(options, apps.SelectOption{
				Value: string(app.Manifest.AppID),
				Label: app.Manifest.DisplayName,
			})
		}
	}
	return apps.NewLookupResponse(options)
}
