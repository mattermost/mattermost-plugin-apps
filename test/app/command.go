package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func commandBindings(cc apps.Context) []apps.Binding {
	b := apps.Binding{
		Label: CommandTrigger,
		Icon:  "icon.png",
		Bindings: []apps.Binding{
			embeddedCommandBinding(cc),
			formCommandBinding(cc),
			subscriptionCommandBinding("subscribe", Subscribe),
			subscriptionCommandBinding("unsubscribe", Unsubscribe),
			numBindingsCommandBinding(cc),
			testCommandBinding(cc),
		},
	}

	return []apps.Binding{b}
}

func numBindingsCommandBinding(_ apps.Context) apps.Binding {
	return apps.Binding{
		Label:       "num_bindings",
		Description: "Choose how many bindings to show in different locations. Provide -1 to use the default options.",
		Icon:        "icon.png",
		Form: &apps.Form{
			Submit: apps.NewCall(NumBindingsPath),
			Fields: []apps.Field{
				{
					Name:                 "location",
					Type:                 apps.FieldTypeStaticSelect,
					Description:          "Location to change",
					IsRequired:           true,
					AutocompletePosition: 1,
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "post_menu",
							Value: "post_menu",
						},
						{
							Label: "channel_header",
							Value: "channel_header",
						},
					},
				},
				{
					Name:                 "number",
					Type:                 apps.FieldTypeText,
					TextSubtype:          apps.TextFieldSubtypeNumber,
					Description:          "Number of bindings to show",
					IsRequired:           true,
					AutocompletePosition: 2,
				},
			},
		},
	}
}

func testCommandBinding(cc apps.Context) apps.Binding {
	out := []apps.Binding{
		{
			Label:    "valid",
			Icon:     "icon.png",
			Bindings: validResponseBindings,
		}, {
			Label:    "error",
			Icon:     "icon.png",
			Bindings: errorResponseBindings,
		}, {
			Label:    "with-input",
			Icon:     "icon.png",
			Bindings: append(validInputBindings, withSubBindings),
		},
	}

	if IncludeInvalid {
		out = append(out,
			apps.Binding{
				Label:    "invalid-response",
				Icon:     "icon.png",
				Bindings: invalidResponseBindings,
			},
			apps.Binding{
				Label:    "invalid-input-binding",
				Icon:     "icon.png",
				Bindings: invalidBindingBindings,
			},
			apps.Binding{
				Label:    "invalid-input-form",
				Icon:     "icon.png",
				Bindings: invalidFormBindings,
			},
		)
	}

	if cc.Channel != nil && cc.Channel.Name == "town-square" {
		out = append([]apps.Binding{
			newBinding("town-square-channel-specific", OK),
		}, out...)
	}

	return apps.Binding{
		Label:    "test-command",
		Bindings: out,
	}
}
