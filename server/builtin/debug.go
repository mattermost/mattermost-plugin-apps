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
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "command.debug.label",
			Other: "debug",
		}),
		Bindings: []apps.Binding{
			a.debugBindingsCommandBinding(loc),
			a.debugCleanCommandBinding(loc),
			{
				Location: "cluster",
				Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.cluster.label",
					Other: "cluster",
				}),
				Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.cluster.description",
					Other: "Test high availability (cluster) functionality.",
				}),
				Bindings: []apps.Binding{
					a.debugClusterTestCachedCommandBinding(loc),
				},
			},
			{
				Location: "kv",
				Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.kv.label",
					Other: "kv",
				}),
				Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
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
			{
				Location: "store",
				Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.store.label",
					Other: "store",
				}),
				Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.store.description",
					Other: "View and update raw plugin KV store.",
				}),
				Bindings: []apps.Binding{
					a.debugStoreListCommandBinding(loc),
					a.debugStorePolluteCommandBinding(loc),
				},
			},
			{
				Location: "sessions",
				Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.label",
					Other: "sessions",
				}),
				Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.session.description",
					Other: "View App specific sessions.",
				}),
				Bindings: []apps.Binding{
					a.debugSessionsListBinding(loc),
					a.debugSessionsViewBinding(loc),
					a.debugSessionsRevokeBinding(loc),
				},
			},
			{
				Location: "oauth",
				Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.oauth.label",
					Other: "oauth",
				}),
				Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
					ID:    "command.debug.oauth.description",
					Other: "View information about the remote OAuth app.",
				}),
				Bindings: []apps.Binding{
					a.debugOAuthConfigViewBinding(loc),
				},
			},
		},
	}
}

func (a *builtinApp) debugIDField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fID,
		Type: apps.FieldTypeText,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.label",
			Other: "id",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.description",
			Other: "App-specific ID, any length.",
		}),
		AutocompleteHint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.id.hint",
			Other: "[ id ]",
		}),
	}
}

func (a *builtinApp) debugBase64KeyField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fBase64Key,
		Type: apps.FieldTypeText,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.label",
			Other: "base64_key",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.description",
			Other: "base64-encoded key, see output of `debug kv list`.",
		}),
		AutocompleteHint: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64key.hint",
			Other: "[ base64-encoded key ]",
		}),
	}
}

func (a *builtinApp) debugBase64Field(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fBase64,
		Type: apps.FieldTypeBool,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64.label",
			Other: "base64",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.kv.base64.description",
			Other: "base64 encoded keys to use in other `debug kv` commands.",
		}),
		Value: true,
	}
}

func (a *builtinApp) debugHashkeysField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fHashkeys,
		Type: apps.FieldTypeBool,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.hashkeys.label",
			Other: "hashkeys",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.count.description",
			Other: "display hashkeys only, parsed.",
		}),
	}
}

func (a *builtinApp) debugCountField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fCount,
		Type: apps.FieldTypeText,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.count.label",
			Other: "count",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.count.description",
			Other: "number of keys per page.",
		}),
	}
}

func (a *builtinApp) debugPageField(loc *i18n.Localizer) apps.Field {
	return apps.Field{
		Name: fPage,
		Type: apps.FieldTypeText,
		Label: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.page.label",
			Other: "page",
		}),
		Description: a.api.I18N.LocalizeDefaultMessage(loc, &i18n.Message{
			ID:    "field.store.page.description",
			Other: "page number.",
		}),
	}
}
