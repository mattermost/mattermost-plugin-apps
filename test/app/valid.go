package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var validResponseBindings = []apps.Binding{
	newBinding("OK", OK),
	newBinding("OK-empty", OKEmpty),
	newBinding("form", FormSimple),
	newBinding("form-source", FormSimpleSource), // does not work, move to invalid?
	newBinding("navigate-external", NavigateExternal),
	newBinding("navigate-internal", NavigateInternal),
}

var errorResponseBindings = []apps.Binding{
	newBinding("error", ErrorDefault),
	newBinding("error-empty", ErrorEmpty),
	newBinding("error-404", Error404),
	newBinding("error-500", Error500),
}

var validInputBindings = []apps.Binding{
	{
		Label:       "empty-form",
		Icon:        "icon.png",
		Description: "Empty submittable form is included in the binding, no flags",
		Form: &apps.Form{
			Submit: callOK,
		},
	},
	{
		Label:       "simple-form",
		Icon:        "icon.png",
		Description: "Simple form is included in the binding",
		Form:        &simpleForm,
	},
	{
		// does not work, move to invalid?
		Label:       "simple-form-source",
		Icon:        "icon.png",
		Description: "Simple form is referenced (`source=`) in the binding, DOES NOT WORK",
		Form:        &simpleFormSource,
	},
	{
		Label:       "full-form",
		Icon:        "icon.png",
		Form:        &fullForm,
		Description: "Full form is included in the binding",
	},
}

var withSubBindings = apps.Binding{
	Label:       "sub-bindings",
	Icon:        "icon.png",
	Description: "Two sub-bindings",
	Bindings: []apps.Binding{
		{
			Label:  "sub1",
			Icon:   "icon.png",
			Submit: callOK,
		},
		{
			Label:  "sub2",
			Icon:   "icon.png",
			Submit: callOK,
		},
	},
}
