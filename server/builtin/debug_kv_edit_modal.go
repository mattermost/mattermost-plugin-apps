// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pkg/errors"
)

func (a *builtinApp) debugKVEditModal() handler {
	return handler{
		requireSysadmin: true,

		formf: func(creq apps.CallRequest) (*apps.Form, error) {
			key, _ := creq.State.(string)
			if key == "" {
				return nil, utils.NewInvalidError(`expected "key" in call State`)
			}

			value := []byte{}
			err := a.conf.MattermostAPI().KV.Get(key, &value)
			if err != nil {
				return nil, err
			}

			loc := a.newLocalizer(creq)
			return &apps.Form{
				Title: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "modal.kv.edit.title",
					Other: "Edit app's KV record",
				}),
				Header: fmt.Sprintf("Key:\n```\n%s\n```\n", key),
				Fields: []apps.Field{
					{
						Name: fCurrentValue,
						ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
							ID:    "field.kv.current_value.modal_label",
							Other: "Current value",
						}),
						Type:        apps.FieldTypeText,
						ReadOnly:    true,
						Value:       string(value),
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name: fNewValue,
						ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
							ID:    "field.kv.new_value.modal_label",
							Other: "New value to save",
						}),
						Type:        apps.FieldTypeText,
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name: fAction,
						ModalLabel: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
							ID:    "field.kv.action.modal_label",
							Other: "Action to take",
						}),
						Type: apps.FieldTypeStaticSelect,
						SelectStaticOptions: []apps.SelectOption{
							{
								Value: "store",
								Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
									ID:    "option.kv.store.label",
									Other: "Store New Value",
								}),
							},
							{
								Value: "delete",
								Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
									ID:    "option.kv.delete.label",
									Other: "Delete Key",
								}),
							},
						},
					},
				},
				SubmitButtons: fAction,
				Call: &apps.Call{
					Path: pDebugKVEditModal,
					Expand: &apps.Expand{
						ActingUser: apps.ExpandSummary,
					},
					State: key,
				},
			}, nil
		},

		submitf: func(creq apps.CallRequest) apps.CallResponse {
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
		},
	}
}
