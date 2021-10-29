// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

var debugKVListCall = apps.Call{
	Path: pDebugKVList,
	Expand: &apps.Expand{
		ActingUser: apps.ExpandSummary,
	},
}

func (a *builtinApp) debugKVList() handler {
	return handler{
		requireSysadmin: true,

		commandBinding: func(loc *i18n.Localizer) apps.Binding {
			return apps.Binding{
				Location:    "list",
				Label:       "list",                                                             // <>/<> Localize
				Hint:        "[ AppID ]",                                                        // <>/<> Localize
				Description: "Display the list of KV keys for an app, in a specific namespace.", // <>/<> Localize
				Call:        &debugKVInfoCall,
				Form:        a.appIDForm(debugKVListCall, loc, namespaceField, base64Field),
			}
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
			appID := apps.AppID(creq.GetValue(fAppID, ""))
			namespace := creq.GetValue(fNamespace, "")
			encode := creq.BoolValue(fBase64)
			app, err := a.proxy.GetInstalledApp(appID)
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			keys := []string{}
			err = a.appservices.KVList(app.BotUserID, namespace, func(key string) error {
				keys = append(keys, key)
				return nil
			})
			if err != nil {
				return apps.NewErrorResponse(err)
			}

			message := fmt.Sprintf("%v total keys for `%s`", len(keys), appID) // <>/<> Localize
			if namespace != "" {
				message += fmt.Sprintf(", namespace `%s`\n", namespace)
			} else {
				message += "\n"
			}
			if encode {
				message += "**NOTE**: keys are base64-encoded for pasting into " +
					"`/apps debug kv edit` command. Use `/apps debug kv list --base64 false` " +
					"to output raw values.\n"
				for _, key := range keys {
					message += fmt.Sprintf("- `%s`\n", base64.URLEncoding.EncodeToString([]byte(key)))
				}
			} else {
				message += "```\n"
				for _, key := range keys {
					message += fmt.Sprintln(key)
				}
				message += "```\n"
			}
			return apps.NewTextResponse(message)
		},

		lookupf: a.debugAppNamespaceLookup,
	}
}

func (a *builtinApp) debugAppNamespaceLookup(creq apps.CallRequest) ([]apps.SelectOption, error) {
	switch creq.SelectedField {
	case fAppID:
		return a.lookupAppID(creq, func(app apps.ListedApp) bool {
			return app.Installed
		})

	case fNamespace:
		return a.lookupNamespace(creq)
	}
	return nil, utils.ErrNotFound
}

func (a *builtinApp) lookupNamespace(creq apps.CallRequest) ([]apps.SelectOption, error) {
	if creq.SelectedField != fNamespace {
		return nil, errors.Errorf("unknown field %q", creq.SelectedField)
	}
	appID := apps.AppID(creq.GetValue(fAppID, ""))
	if appID == "" {
		return nil, errors.Errorf("please select --" + fAppID + " first")
	}

	var options []apps.SelectOption
	_, namespaces, err := a.debugListKeys(appID)
	if err != nil {
		return nil, err
	}
	for ns, c := range namespaces {
		options = append(options, apps.SelectOption{
			Value: ns,
			Label: fmt.Sprintf("%q (%v keys)", ns, c),
		})
	}
	return options, nil
}
