// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

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
				Location: "list",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.list.label",
					Other: "list",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.list.description",
					Other: "Display the list of KV keys for an app, in a specific namespace.",
				}),
				Hint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.list.hint",
					Other: "[ AppID Namespace ]",
				}),
				Call: &debugKVInfoCall,
				Form: a.appIDForm(debugKVListCall, loc,
					a.debugNamespaceField(loc), a.debugBase64Field(loc)),
			}
		},

		submitf: func(_ context.Context, creq apps.CallRequest) apps.CallResponse {
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

			loc := a.newLocalizer(creq)
			message := a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "command.debug.kv.list.submit.message",
					Other: "{{.Count}} total keys for `{{.AppID}}`",
				},
				TemplateData: map[string]string{
					"Count": strconv.Itoa(len(keys)),
					"AppID": string(appID),
				},
			})

			if namespace != "" {
				message += a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "command.debug.kv.list.submit.namespace",
						Other: ", namespace `{{.Namespace}}`\n",
					},
					TemplateData: map[string]string{
						"Namespace": namespace,
					},
				})
			} else {
				message += "\n"
			}
			if encode {
				message += a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.list.submit.note",
					Other: "**NOTE**: keys are base64-encoded for pasting into `/apps debug kv edit` command. Use `/apps debug kv list --base64 false` to output raw values.\n",
				})
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
