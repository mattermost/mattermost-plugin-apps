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
			timerCommandBinding("timer-create", CreateTimer),

			testCommandBinding(cc),
		},
	}

	return []apps.Binding{b}
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
