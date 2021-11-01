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
				Location:    "info",
				Label:       a.conf.Local(loc, "command.debug.kv.info.label"),
				Description: a.conf.Local(loc, "command.debug.kv.info.description"),
				Hint:        a.conf.Local(loc, "command.debug.kv.info.hint"),
				Call:        &debugKVInfoCall,
				Form:        a.appIDForm(debugKVInfoCall, loc),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			n, namespaces, err := a.debugListKeys(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}
			loc := a.newLocalizer(creq)

			message := a.conf.LocalWithTemplate(loc, "command.debug.kv.info.submit.message",
				map[string]string{
					"Count": strconv.Itoa(n),
					"AppID": string(appID),
				})
			if len(namespaces) > 0 {
				message += a.conf.Local(loc, "command.debug.kv.info.submit.namespace")
			}
			for ns, c := range namespaces {
				if ns == "" {
					ns = a.conf.Local(loc, "command.debug.kv.info.submit.none")
				}
				message += fmt.Sprintf("  - `%s`: %v\n", ns, c)
			}
			return apps.NewTextResponse(message)
		},

		lookupf: func(creq apps.CallRequest) ([]apps.SelectOption, error) {
			return a.lookupAppID(creq, nil)
		},
	}
}

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
		namespaces[ns] = namespaces[ns] + 1
		n++
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return n, namespaces, nil
}
