// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
)

func (a *builtinApp) debugKVListCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "list",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.list.label",
			Other: "list",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.list.description",
			Other: "Display the list of KV keys for an app, in a specific namespace.",
		}),
		Hint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.list.hint",
			Other: "[ AppID Namespace ]",
		}),
		Form: &apps.Form{
			Submit: newUserCall(PathDebugKVList),
			Fields: []apps.Field{
				a.appIDField(LookupInstalledApps, 1, true, loc),
				a.namespaceField(0, false, loc),
				a.debugBase64Field(loc),
			},
		},
	}
}

func (a *builtinApp) debugKVList(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	appID := apps.AppID(creq.GetValue(FieldAppID, ""))
	namespace := creq.GetValue(FieldNamespace, "")
	encode := creq.BoolValue(fBase64)

	keys := []string{}
	appservicesRequest := r.WithSourceAppID(appID)
	err := a.appservices.KVList(appservicesRequest, namespace, func(key string) error {
		keys = append(keys, key)
		return nil
	})
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	loc := a.newLocalizer(creq)
	message := a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
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
		message += a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "command.debug.kv.list.submit.namespace",
				Other: ", namespace `{{.Namespace}}`",
			},
			TemplateData: map[string]string{
				"Namespace": namespace,
			},
		})
	}
	message += "\n"

	if encode {
		message += a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.kv.list.submit.note",
			Other: "**NOTE**: keys are base64-encoded for pasting into `/apps debug kv edit` command. Use `/apps debug kv list --base64 false` to output raw values.",
		}) + "\n"
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

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: keys,
		Text: message,
	}
}
