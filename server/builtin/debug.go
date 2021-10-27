// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package builtin

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func (a *builtinApp) debugCommandBinding() apps.Binding {
	return apps.Binding{
		Label:    "debug",
		Location: "debug",
		Bindings: []apps.Binding{
			a.debugBindings().commandBinding(),
			a.debugClean().commandBinding(),
			{
				Label:       "kv",
				Location:    "kv",
				Description: "View and update apps' KV stores.",
				Bindings: []apps.Binding{
					a.debugKVClean().commandBinding(),
					a.debugKVEdit().commandBinding(),
					a.debugKVInfo().commandBinding(),
					a.debugKVList().commandBinding(),
				},
			},
		},
	}
}

var namespaceField = apps.Field{
	Name:             fNamespace,
	Label:            fNamespace,
	Type:             apps.FieldTypeDynamicSelect,
	Description:      "Select a namespace, see `debug kv info` for the list of app's namespaces.",
	AutocompleteHint: "[ namespace ]",
}

var base64Field = apps.Field{
	Name:        fBase64,
	Label:       fBase64,
	Type:        apps.FieldTypeBool,
	Description: "base64-encode keys for pasting into `/apps debug kv edit` command.",
	Value:       true,
}
