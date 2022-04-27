package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

var lookupForm = apps.Form{
	Title:  "Test Lookup Form",
	Submit: callOK,
	Fields: []apps.Field{
		{
			Name:                "simple",
			Type:                apps.FieldTypeDynamicSelect,
			SelectDynamicLookup: apps.NewCall(Lookup),
		},
		{
			Name:                "multiword",
			Type:                apps.FieldTypeDynamicSelect,
			SelectDynamicLookup: apps.NewCall(LookupMultiword),
		},
		{
			Name:                "empty",
			Type:                apps.FieldTypeDynamicSelect,
			SelectDynamicLookup: apps.NewCall(LookupEmpty),
		},
		{
			Name:                "invalid",
			Type:                apps.FieldTypeDynamicSelect,
			SelectDynamicLookup: apps.NewCall(InvalidLookup),
		},
	},
}
