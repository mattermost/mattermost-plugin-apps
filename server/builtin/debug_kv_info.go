// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugKVInfoCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "info",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.label",
			Other: "info",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.description",
			Other: "Display KV store statistics for an app..",
		}),
		Hint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.hint",
			Other: "[ AppID ]",
		}),
		Form: &apps.Form{
			Submit: newUserCall(PathDebugKVInfo),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 0, false, loc),
			},
		},
	}
}

func (a *builtinApp) debugKVInfo(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	if appID == "" {
		return a.debugKVInfoForAll(r, creq)
	}
	return a.debugKVInfoForApp(r, creq, appID)
}

func (a *builtinApp) debugKVInfoForAll(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	info, err := a.appservices.KVDebugInfo(r)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	loc := a.newLocalizer(creq)

	message := a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: "command.debug.kv.info.submit.message",
			Other: `{{.Total}} total keys:
- Special cached keys: {{.CachedStoreTotal}}
{{range $name, $count := .CachedStoreCountByName}}
  - ` + "`{{$name}}`" + `: {{$count}}
{{end}}
- OAuth2 temporary state records: {{.OAuth2StateCount}}
- Other internal records: {{.Other}}


- Apps' proprietary KV records: {{.AppsTotal}}
`},
		TemplateData: *info,
	})

	for appID, appInfo := range info.Apps {
		message += fmt.Sprintf("  - `%s`: %v (%v kv, %v users, %v tokens)\n", appID, appInfo.Total(), appInfo.AppKVCount, appInfo.UserCount, appInfo.TokenCount)
	}

	totalKnown := info.OAuth2StateCount + info.AppsTotal + info.Other + info.CachedStoreTotal
	if totalKnown != info.Total {
		message += fmt.Sprintf("- **UNKNOWN**: %v\n", info.Total-totalKnown)
	}

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Text: message + "\n\n" + utils.JSONBlock(info),
		Data: info,
	}
}

func (a *builtinApp) debugKVInfoForApp(r *incoming.Request, creq apps.CallRequest, appID apps.AppID) apps.CallResponse {
	appInfo, err := a.appservices.KVDebugAppInfo(r, appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	loc := a.newLocalizer(creq)

	message := a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.kv.info.submit.message",
			Other: "{{.Count}} total keys for `{{.AppID}}`.",
		},
		TemplateData: map[string]string{
			"AppID": string(appID),
			"Count": strconv.Itoa(appInfo.AppKVCount),
		},
	}) + "\n"

	if len(appInfo.AppKVCountByNamespace) > 0 {
		message += "\n" +
			a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.kv.info.submit.namespaces",
				Other: "Namespaces:",
			}) +
			"\n"
	}
	for ns, c := range appInfo.AppKVCountByNamespace {
		if ns == "" {
			ns = a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
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
