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
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
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
					a.debugKVClean().commandBinding(loc),
					a.debugKVCreate().commandBinding(loc),
					a.debugKVEdit().commandBinding(loc),
					a.debugKVInfo().commandBinding(loc),
					a.debugKVList().commandBinding(loc),
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

func (a *builtinApp) debugNamespaceField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fNamespace,
		Type: apps.FieldTypeDynamicSelect,
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.label",
			Other: "namespace",
		}),
		Description: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.description",
			Other: "App-specific namespace (up to 2 letters). See `debug kv info` for the list of app's namespaces.",
		}),
		AutocompleteHint: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.namespace.hint",
			Other: "namespace (up to 2 letters)",
		}),
	}
}

func (a *builtinApp) debugBase64Field(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fBase64,
		Type: apps.FieldTypeBool,
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
