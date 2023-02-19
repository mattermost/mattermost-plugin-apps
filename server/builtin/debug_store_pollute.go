// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/server/store"
)

func (a *builtinApp) debugStorePolluteCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "list",
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.store.pollute.label",
			Other: "pollute",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.store.pollute.description",
			Other: "Add garbage records to the store.",
		}),
		Form: &apps.Form{
			Submit: newUserCall(PathDebugStorePollute),
			Fields: []apps.Field{
				a.debugCountField(loc),
			},
		},
	}
}

func (a *builtinApp) debugStorePollute(r *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	var err error
	c := 1000
	countStr := creq.GetValue(fCount, "")
	if countStr != "" {
		c, err = strconv.Atoi(countStr)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
	}

	keys := []string{}
	for i := 0; i < c; i++ {
		key := fmt.Sprintf("%s-%v-%d", store.DebugPrefix, time.Now().UnixMilli(), i)
		_, err = a.api.Mattermost.KV.Set(key, []byte("garbage"))
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		keys = append(keys, key)
	}

	loc := a.newLocalizer(creq)
	message := a.api.I18N.LocalizeWithConfig(loc, &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "command.debug.store.pollute.submit.message",
			Other: "Created {{.Count}} garbage keys",
		},
		TemplateData: map[string]string{
			"Count": strconv.Itoa(len(keys)),
		},
	})

	return apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: keys,
		Text: message,
	}
}
