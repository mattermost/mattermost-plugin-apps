// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/incoming"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugKVEditModalForm(_ *incoming.Request, creq apps.CallRequest) apps.CallResponse {
	key, _ := creq.State.(string)
	if key == "" {
		return apps.NewErrorResponse(utils.NewInvalidError(`expected "key" in call State`))
	}

	value := []byte{}
	err := a.conf.MattermostAPI().KV.Get(key, &value)
	if err != nil {
		return apps.NewErrorResponse(err)
	}

	loc := a.newLocalizer(creq)

	buttons := []apps.SelectOption{
		{
			Value: "store",
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "option.kv.store.label",
				Other: "Store New Value",
			}),
		},
	}
	if len(value) > 0 {
		buttons = append(buttons, apps.SelectOption{
			Value: "delete",
			Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
				ID:    "option.kv.delete.label",
				Other: "Delete Key",
			}),
		})
	}

	return apps.NewFormResponse(apps.Form{
		Title: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "modal.kv.edit.title",
			Other: "Edit app's KV record",
		}),
		Header: fmt.Sprintf("Key:\n```\n%s\n```\n", key),
		Fields: []apps.Field{
			{
				Name:        fCurrentValue,
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "field.kv.current_value.modal_label",
					Other: "Current value",
				}),
				ReadOnly: true,
				Value:    string(value),
			},
			{
				Name:        fNewValue,
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "field.kv.new_value.modal_label",
					Other: "New value to save",
				}),
			},
			{
				Name: fAction,
				Type: apps.FieldTypeStaticSelect,
				ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "field.kv.action.modal_label",
					Other: "Action to take",
				}),
				SelectStaticOptions: buttons,
			},
		},
		SubmitButtons: fAction,
		Submit:        newUserCall(pDebugKVEditModal).WithState(key),
	})
}
