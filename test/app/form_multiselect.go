package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var formMultiselect = apps.Form{
	Title:  "Test Multiselect Form",
	Header: "Test header",
	Submit: callOK,
	Fields: []apps.Field{
		{
			Name:          "static",
			Type:          apps.FieldTypeStaticSelect,
			Label:         "static",
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
					Label: "static value 4",
					Value: "sv4",
				},
			},
		},
		{
			Name:          "user",
			Type:          apps.FieldTypeUser,
			Label:         "user",
			SelectIsMulti: true,
		},
		{
			Name:          "channel",
			Type:          apps.FieldTypeChannel,
			Label:         "channel",
			SelectIsMulti: true,
		},
	},
}
