package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var invalidResponseBindings = []apps.Binding{
	newBinding("invalid-navigate", InvalidNavigate),
	newBinding("invalid-form", InvalidForm),
	newBinding("unknown-type", InvalidUnknownType),
	newBinding("HTML-random", InvalidHTML),
	newBinding("JSON-random", ManifestPath),
}

var invalidBindingBindings = []apps.Binding{
	{
		Label:       "valid",
		Description: "the only valid binding in this submenu",
		Icon:        "icon.png",
		Submit:      callOK,
	},
	{
		Label:       "conflicting-label",
		Description: "2 sub-bindings have label `Command`, only one should appear",
		Bindings: []apps.Binding{
			{
				Location: "command1",
				Label:    "Command",
				Submit:   callOK,
			},
			{
				Location: "command2",
				Label:    "Command",
				Submit:   apps.NewCall(ErrorDefault),
			},
		},
	},
	{
		Label:       "space-in-label",
		Description: "`Command with space` is not visible.",
		Bindings: []apps.Binding{
			{
				Label:  "Command with space",
				Submit: apps.NewCall(ErrorDefault),
			},
			{
				Label:  "Command-with-no-space",
				Submit: callOK,
			},
		},
	},
}

var invalidFormBindings = []apps.Binding{
	{
		Label:       "unsubmittable",
		Icon:        "icon.png",
		Description: "Form is included in the binding does not have submit",
		Form: &apps.Form{
			Title: "unsubmittable form",
		},
	},
	{
		Label:       "conflicting-fields",
		Icon:        "icon.png",
		Description: "2 fields have label `field`, only `field1` should appear",
		Form: &apps.Form{
			Submit: callOK,
			Fields: []apps.Field{
				{
					Type:  apps.FieldTypeText,
					Name:  "field1",
					Label: "field",
				},
				{
					Type:  apps.FieldTypeText,
					Name:  "field2",
					Label: "field",
				},
			},
		},
	},
	{
		Label:       "conflicting-options",
		Icon:        "icon.png",
		Description: "2 select options have value `opt`, only `opt1` should appear",
		Form: &apps.Form{
			Submit: callOK,
			Fields: []apps.Field{
				{
					Type: apps.FieldTypeStaticSelect,
					Name: "field",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "opt1",
							Value: "opt",
						},
						{
							Label: "opt2",
							Value: "opt",
						},
					},
				},
			},
		},
	},
	{
		Label:       "empty-option",
		Icon:        "icon.png",
		Description: "a select option has no name/value; only `opt1` should appear",
		Form: &apps.Form{
			Submit: callOK,
			Fields: []apps.Field{
				{
					Type: apps.FieldTypeStaticSelect,
					Name: "field",
					SelectStaticOptions: []apps.SelectOption{
						{
							Label: "opt1",
							Value: "opt",
						},
						{
							Label: "",
							Value: "",
						},
					},
				},
			},
		},
	},
	{
		Label:       "form-space-in-field-label",
		Icon:        "icon.png",
		Description: "a form field has a label with a space, only `field-with-no-space` should appear.",
		Form: &apps.Form{
			Submit: callOK,
			Fields: []apps.Field{
				{
					Type:  apps.FieldTypeText,
					Name:  "field1",
					Label: "field with space",
				},
				{
					Type:  apps.FieldTypeText,
					Name:  "field2",
					Label: "field-with-no-space",
				},
			},
		},
	},
	{
		Label:       "empty-lookup",
		Description: "a form field is a dynamic select with an empty lookup response.",
		Icon:        "icon.png",
		Form: &apps.Form{
			Submit: apps.NewCall(ErrorDefault),
			Fields: []apps.Field{
				{
					Type:                 apps.FieldTypeDynamicSelect,
					IsRequired:           true,
					Name:                 "field1",
					AutocompletePosition: 1,
					SelectDynamicLookup:  apps.NewCall(LookupEmpty),
				},
			},
		},
	},
}
