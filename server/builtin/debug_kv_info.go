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
				a.appIDField(LookupInstalledApps, 1, false, loc),
			},
		},
	}
}

func (a *builtinApp) debugKVInfo(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	if appID != "" {
		return a.debugKVInfoForApp(r, creq, appID)
	}
	return a.debugKVInfoForAll(r, creq)
}

func (a *builtinApp) debugKVInfoForAll(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	info, err := a.appservices.KVDebugInfo(r)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	loc := a.newLocalizer(creq)

	message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: "command.debug.kv.info.submit.message",
			Other: `{{.Total}} total keys:
- Manifest records: {{.Manifests}}
- App records: {{.Apps}}
- Subscription records: {{.Subscriptions}}
- OAuth2 temporary state records: {{.OAuth2State}}
- Other internal records: {{.Other}}
- Apps' own records: {{.AppsTotal}}
`},
		TemplateData: map[string]string{
			"Total":         strconv.Itoa(info.Total),
			"Manifests":     strconv.Itoa(info.ManifestCount),
			"Apps":          strconv.Itoa(info.InstalledAppCount),
			"Subscriptions": strconv.Itoa(info.SubscriptionCount),
			"OAuth2State":   strconv.Itoa(info.OAuth2StateCount),
			"Other":         strconv.Itoa(info.Other),
			"AppsTotal":     strconv.Itoa(info.AppsTotal),
		},
	})

	for appID, appInfo := range info.Apps {
		message += fmt.Sprintf("  - `%s`: %v (%v kv, %v users, %v tokens)\n", appID, appInfo.Total(), appInfo.AppKVCount, appInfo.UserCount, appInfo.TokenCount)
	}

	totalKnown := info.ManifestCount + info.InstalledAppCount + info.SubscriptionCount + info.OAuth2StateCount + info.AppsTotal + info.Other
	if totalKnown != info.Total {
		message += fmt.Sprintf("- **UNKNOWN**: %v\n", info.Total-totalKnown)
	}

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Text: message,
		Data: info,
	}
}

func (a *builtinApp) debugKVInfoForApp(r *incoming.Request, creq apps.CallRequest, appID apps.AppID) apps.CallResponse {
	appInfo, err := a.appservices.KVDebugAppInfo(r, appID)
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
			"Count": strconv.Itoa(appInfo.AppKVCount),
		},
	}) + "\n"

	if len(appInfo.AppKVCountByNamespace) > 0 {
		message += "\n" +
			a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.kv.info.submit.namespaces",
				Other: "Namespaces:",
			}) +
			"\n"
	}
	for ns, c := range appInfo.AppKVCountByNamespace {
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
