// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils"
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
				Title:  a.conf.Local(loc, "modal.kv.edit.title"),
				Header: fmt.Sprintf("Key:\n```\n%s\n```\n", key),
				Fields: []apps.Field{
					{
						Name:        fCurrentValue,
						ModalLabel:  a.conf.Local(loc, "field.kv.current_value.modal_label"),
						Type:        apps.FieldTypeText,
						ReadOnly:    true,
						Value:       string(value),
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name:        fNewValue,
						ModalLabel:  a.conf.Local(loc, "field.kv.new_value.modal_label"),
						Type:        apps.FieldTypeText,
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name:       fAction,
						ModalLabel: a.conf.Local(loc, "field.kv.action.modal_label"),
						Type:       apps.FieldTypeStaticSelect,
						SelectStaticOptions: []apps.SelectOption{
							{
								Value: "store",
								Label: a.conf.Local(loc, "option.kv.store.label"),
							},
							{
								Value: "delete",
								Label: a.conf.Local(loc, "option.kv.delete.label"),
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
					a.conf.LocalWithTemplate(loc, "modal.kv.edit.submit.stored",
						map[string]string{
							"Key":   key,
							"Value": newValue,
						}))

			case "delete":
				err := mm.KV.Delete(key)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				return apps.NewTextResponse(
					a.conf.LocalWithTemplate(loc, "modal.kv.edit.submit.deleted",
						map[string]string{
							"Key":   key,
							"Value": newValue,
						}))

			default:
				return apps.NewErrorResponse(errors.New("don't know what to do: %q"))
			}
		},
	}
}
