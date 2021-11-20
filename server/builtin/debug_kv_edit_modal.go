// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
)

func (a *builtinApp) debugKVEditModal(creq apps.CallRequest) apps.CallResponse {
	action := creq.GetValue(fAction, "")
	newValue := creq.GetValue(fNewValue, "")
	key, _ := creq.State.(string)
	loc := a.newLocalizer(creq)

	mm := a.conf.MattermostAPI()
	switch action {
	case "store":
		_, err := mm.KV.Set(key, []byte(newValue))
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		return apps.NewTextResponse(
			a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.kv.edit.submit.stored",
					Other: "Stored:\n```\nKey: {{.Key}}\n\n{{.Value}}\n```\n",
				},
				TemplateData: map[string]string{
					"Key":   key,
					"Value": newValue,
				},
			}))

	case "delete":
		err := mm.KV.Delete(key)
		if err != nil {
			return apps.NewErrorResponse(err)
		}
		return apps.NewTextResponse(
			a.conf.I18N().LocalizeWithConfig(loc, &i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "modal.kv.edit.submit.deleted",
					Other: "Deleted:\n```\nKey: {{.Key}}\n```\n",
				},
				TemplateData: map[string]string{
					"Key": key,
				},
			}))

	default:
		return apps.NewErrorResponse(errors.New("don't know what to do: %q"))
	}
}

func (a *builtinApp) debugKVEditModalForm(creq apps.CallRequest) apps.CallResponse {
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
				SelectOptions: buttons,
			},
		},
		SubmitButtons: fAction,
		Submit:        newUserCall(pDebugKVEditModal).WithState(key),
	})
}
