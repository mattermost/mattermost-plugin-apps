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

func (a *builtinApp) debugStoreListCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "list",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.store.list.label",
			Other: "list",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.store.list.description",
			Other: "Display the list of all raw keys in the store.",
		}),
		Form: &apps.Form{
			Submit: newUserCall(PathDebugStoreList),
			Fields: []apps.Field{
				a.debugCountField(loc),
				a.debugHashkeysField(loc),
				a.debugPageField(loc),
			},
		},
	}
}

func (a *builtinApp) debugStoreList(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	var err error
	page := 0
	pageStr := creq.GetValue(fPage, "")
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
	}

	count := store.ListKeysPerPage
	countStr := creq.GetValue(fCount, "")
	if countStr != "" {
		count, err = strconv.Atoi(countStr)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
	}

	keys, err := a.api.Mattermost.KV.ListKeys(page, count)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	loc := a.newLocalizer(creq)
	message := a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.store.list.submit.message",
			Other: "{{.Count}} keys",
		},
		TemplateData: map[string]string{
			"Count": strconv.Itoa(len(keys)),
		},
	})
	message += "\n\n"

	if creq.BoolValue(fHashkeys) {
		message += "| prefix | app ID | user ID | ns  | key id (hash) |\n"
		message += "| :------| :----- | :------ | :-- | :------------ |\n"

		for _, key := range keys {
			prefix, appID, userID, namespace, idhash, parseErr := store.ParseHashkey(key)
			if parseErr != nil {
				continue
			}
			message += fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s` |\n", prefix, appID, userID, namespace, idhash)
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
