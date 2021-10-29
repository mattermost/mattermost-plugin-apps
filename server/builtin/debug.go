// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func (a *builtinApp) debugCommandBinding(loc *i18n.Localizer) apps.Binding {
	return apps.Binding{
		Label: a.conf.I18N().LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.label",
			Other: "debug",
		}),
		Location: "debug",
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(loc),
			a.debugClean().commandBinding(loc),
			{
				Label:       "kv",                               // <>/<> TODO localize
				Location:    "kv",                               // <>/<> TODO localize
				Description: "View and update apps' KV stores.", // <>/<> TODO localize
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

var namespaceField = apps.Field{
	Name:             fNamespace,
	Label:            fNamespace, // <>/<> TODO localize
	Type:             apps.FieldTypeDynamicSelect,
	Description:      "Select a namespace, see `debug kv info` for the list of app's namespaces.", // <>/<> TODO localize
	AutocompleteHint: "[ namespace ]",                                                             // <>/<> TODO localize
}

var base64Field = apps.Field{
	Name:        fBase64,
	Label:       fBase64, // <>/<> TODO localize
	Type:        apps.FieldTypeBool,
	Description: "base64-encode keys for pasting into `/apps debug kv edit` command.", // <>/<> TODO localize
	Value:       true,
}
