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

			return &apps.Form{
				Title:  "Edit KV record",
				Header: fmt.Sprintf("Key:\n```\n%s\n```\n", key),
				Fields: []apps.Field{
					{
						Name:        fCurrentValue,
						ModalLabel:  "Current value",
						Type:        apps.FieldTypeText,
						ReadOnly:    true,
						Value:       string(value),
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name:        fNewValue,
						ModalLabel:  "New value to save",
						Type:        apps.FieldTypeText,
						TextSubtype: apps.TextFieldSubtypeTextarea,
					},
					{
						Name:       fAction,
						ModalLabel: "Action to take",
						Type:       apps.FieldTypeStaticSelect,
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "Store New Value",
								Value: "store",
							},
							{
								Label: "Delete Key",
								Value: "delete",
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

			mm := a.conf.MattermostAPI()
			switch action {
			case "store":
				_, err := mm.KV.Set(key, []byte(newValue))
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				return apps.NewTextResponse("Stored:\n```\nKey: %s\n\n%s\n```\n", key, newValue)

			case "delete":
				err := mm.KV.Delete(key)
				if err != nil {
					return apps.NewErrorResponse(err)
				}
				return apps.NewTextResponse("Deleted:\n```\nKey: %s\n```\n", key)

			default:
				return apps.NewErrorResponse(errors.New("don't know what to do: %q"))
			}
		},
	}
}
