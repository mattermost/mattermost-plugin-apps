// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) debugKVInfoCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "info",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.label",
			Other: "info",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.description",
			Other: "Display KV store statistics for an app..",
		}),
		Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.hint",
			Other: "[ AppID ]",
		}),
		Form: &apps.Form{
			Submit: newUserCall(PathDebugKVInfo),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) debugKVInfo(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	appInfo, err := a.appservices.KVDebugAppInfo(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	loc := a.newLocalizer(creq)

	message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.kv.info.submit.message",
			Other: "{{.Count}} total keys for `{{.AppID}}`.",
		},
		TemplateData: map[string]string{
			"AppID": string(appID),
			"Count": strconv.Itoa(appInfo.AppCount),
		},
	}) + "\n"

	if len(appInfo.AppByNamespace) > 0 {
		message += "\n" +
			a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.kv.info.submit.namespaces",
				Other: "Namespaces:",
			}) +
			"\n"
	}
	for ns, c := range appInfo.AppByNamespace {
		if ns == "" {
			ns = a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.kv.info.submit.none",
				Other: "(none)",
			})
		}
		message += fmt.Sprintf("  - `%s`: %v\n", ns, c)
	}

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Text: message,
		Data: appInfo,
	}
}
