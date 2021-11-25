// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

var debugKVInfoCall = apps.Call{
	Path: pDebugKVInfo,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVInfo() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
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
				Call: &debugKVInfoCall,
				Form: a.appIDForm(debugKVInfoCall, loc),
			}
		},

		submitf: func(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			r.SetAppID(appID)
			n, namespaces, err := a.debugListKeys(r, appID)
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
		},

		lookupf: func(r *incoming.Request, creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(r, creq, nil)
		},
	}
}

func (a *builtinApp) debugListKeys(r *incoming.Request, appID apps.AppID) (int, map[string]int, error) {
	n := 0
	namespaces := map[string]int{}
	err := a.appservices.KVList(r,
		appID, r.ActingUserID(),
		"", func(key string) error {
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
