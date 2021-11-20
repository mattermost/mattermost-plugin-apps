// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) debugCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Location: "debug",
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.label",
			Other: "debug",
		}),
		Bindings: []apps.Binding{
			a.debugBindingsCommandBinding(loc),
			a.debugCleanCommandBinding(loc),
			{
				Location: "kv",
				Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.label",
					Other: "kv",
				}),
				Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.description",
					Other: "View and update apps' KV stores.",
				}),
				Bindings: []apps.Binding{
					a.debugKVCleanCommandBinding(loc),
					a.debugKVCreateCommandBinding(loc),
					a.debugKVEditCommandBinding(loc),
					a.debugKVInfoCommandBinding(loc),
					a.debugKVListCommandBinding(loc),
				},
			},
		},
	}
}

func (a *builtinApp) debugIDField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fID,
		Type: apps.FieldTypeText,
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.label",
			Other: "id",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.description",
			Other: "App-specific ID, any length.",
		}),
		AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.hint",
			Other: "[ id ]",
		}),
	}
}

func (a *builtinApp) debugBase64Field(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fBase64,
		Type: apps.FieldTypeBool,
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64.label",
			Other: "base64",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64.description",
			Other: "base64-encode keys, so they can be cut-and-pasted.",
		}),
		Value: true,
	}
}

func (a *builtinApp) debugBase64KeyField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fBase64Key,
		Type: apps.FieldTypeText,
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.label",
			Other: "base64_key",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.description",
			Other: "base64-encoded key, see output of `debug kv list`.",
		}),
		AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.hint",
			Other: "[ base64-encoded key ]",
		}),
		Value: true,
	}
}
