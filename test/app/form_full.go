package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var fullForm = apps.Form{
	Title:  "Test Full Form",
	Header: "Test header",
	Submit: callOK,
	Fields: []apps.Field{
		{
			Name: "lookup",
			Type: apps.FieldTypeDynamicSelect,

			SelectDynamicLookup: apps.NewCall(Lookup),
		},
		{
			Name: "text",
			Type: apps.FieldTypeText,
		},
		{
			Type: "markdown",
			Name: string(apps.FieldTypeMarkdown),

			Description: "***\n## User information\nRemember to fill all these fields with the **user** information, not the general information.",
		},
		{
			Name: "boolean",
			Type: apps.FieldTypeBool,
		},
		{
			Name: "channel",
			Type: apps.FieldTypeChannel,
		},
		{
			Name: "user",
			Type: apps.FieldTypeUser,
		},
		{
			Name: "static",
			Type: apps.FieldTypeStaticSelect,

			SelectStaticOptions: []apps.SelectOption{
				{
					Label: "static value 1",
					Value: "sv1",
				},
				{
					Label: "static value 2",
					Value: "sv2",
				},
			},
		},
		{
			Name: "multi",
			Type: apps.FieldTypeStaticSelect,

			SelectIsMulti: true,
			SelectStaticOptions: []apps.SelectOption{
				{
					Label: "static value 1",
					Value: "sv1",
				},
				{
					Label: "static value 2",
					Value: "sv2",
				},
				{
					Label: "static value 3",
					Value: "sv3",
				},
				{
					Label: "1",
					Value: "1",
				},
				{
					Label: "2",
					Value: "2",
				},
				{
					Label: "3",
					Value: "3",
				},
				{
					Label: "4",
					Value: "4",
				},
				{
					Label: "5",
					Value: "5",
				},
				{
					Label: "6",
					Value: "6",
				},
				{
					Label: "7",
					Value: "7",
				},
				{
					Label: "8",
					Value: "8",
				},
				{
					Label: "9",
					Value: "9",
				},
				{
					Label: "10",
					Value: "10",
				},
			},
		},
		{
			Name:     "user_readonly",
			Type:     apps.FieldTypeUser,
			ReadOnly: true,
			Value: apps.SelectOption{
				Label: "anne.stone",
				Value: "hyspg5mhapgffjnmzdcmko4qzw",
			},
		},
		{
			Name:     "static_readonly",
			Type:     apps.FieldTypeStaticSelect,
			ReadOnly: true,
			SelectStaticOptions: []apps.SelectOption{
				{
					Label: "static value 1",
					Value: "sv1",
				},
				{
					Label: "static value 2",
					Value: "sv2",
				},
			},
			Value: apps.SelectOption{
				Label: "static value 2",
				Value: "sv2",
			},
		},
	},
}

var fullFormReadonly = func() apps.Form {
	var fields []apps.Field
	for _, f := range fullForm.Fields {
		f.ReadOnly = true
		fields = append(fields, f)
	}

	form := fullForm
	form.Fields = fields
	return form
}
