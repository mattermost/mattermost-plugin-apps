// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func (a *builtinApp) debugListKeys(appID apps.AppID) (int, map[string]int, error) {
	app, err := a.proxy.GetInstalledApp(appID)
	if err != nil {
		return 0, nil, err
	}

	n := 0
	namespaces := map[string]int{}
	err = a.appservices.KVList(app.BotUserID, "", func(key string) error {
		_, _, ns, _, e := store.ParseHashkey(key)
		if e != nil {
			return e
		}
		namespaces[ns]++
		n++
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return n, namespaces, nil
}

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
			Submit: newUserCall(pDebugKVInfo),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
			},
		},
	}
}

func (a *builtinApp) debugKVInfo(creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	n, namespaces, err := a.debugListKeys(appID)
	if err != nil {
		return apps.NewErrorResponse(err)
	}
	loc := a.newLocalizer(creq)

	message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.kv.info.submit.message",
			Other: "{{.Count}} total keys for `{{.AppID}}`.\n",
		},
		TemplateData: map[string]string{
			"Count": strconv.Itoa(n),
			"AppID": string(appID),
		},
	})

	if len(namespaces) > 0 {
		message += a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.info.submit.namespaces",
			Other: "\nNamespaces:\n",
		})
	}
	for ns, c := range namespaces {
		if ns == "" {
			ns = a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "command.debug.kv.info.submit.none",
				Other: "(none)",
			})
		}
		message += fmt.Sprintf("  - `%s`: %v\n", ns, c)
	}
	return apps.NewTextResponse(message)
}
