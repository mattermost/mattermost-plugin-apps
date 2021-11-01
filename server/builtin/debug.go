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
		Label:    a.conf.Local(loc, "command.debug.label"),
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
			{
				Location:    "kv",
				Label:       a.conf.Local(loc, "command.debug.kv.label"),
				Description: a.conf.Local(loc, "command.debug.kv.description"),
				Bindings: []apps.Binding{
					a.debugKVClean().commandBinding(loc),
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
		Name:             fID,
		Type:             apps.FieldTypeText,
		Label:            a.conf.Local(loc, "field.kv.id.label"),
		Description:      a.conf.Local(loc, "field.kv.id.description"),
		AutocompleteHint: a.conf.Local(loc, "field.kv.id.hint"),
	}
}

func (a *builtinApp) debugNamespaceField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name:             fNamespace,
		Type:             apps.FieldTypeDynamicSelect,
		Label:            a.conf.Local(loc, "field.kv.namespace.label"),
		Description:      a.conf.Local(loc, "field.kv.namespace.description"),
		AutocompleteHint: a.conf.Local(loc, "field.kv.namespace.hint"),
	}
}

func (a *builtinApp) debugBase64Field(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name:             fBase64,
		Type:             apps.FieldTypeBool,
		Label:            a.conf.Local(loc, "field.kv.base64key.label"),
		Description:      a.conf.Local(loc, "field.kv.base64key.description"),
		AutocompleteHint: a.conf.Local(loc, "field.kv.base64key.hint"),
		Value:            true,
	}
}
