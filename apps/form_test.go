// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestUnmarshalForm(t *testing.T) {
	const full = `
	{
		"source": "/sourcepath",
		"title": "form title",
		"header": "form header",
		"footer": "form footer",
		"icon": "/iconpath",
		"submit": {
			"path": "/submitpath"
		},
		"submit_buttons": "fieldName",
		"fields": [
			{
				"name": "field1",
				"type": "text"
			},
			{
				"name": "field2",
				"type": "text"
			}
		]
	}
	`

	f := apps.Form{}
	err := json.Unmarshal([]byte(full), &f)
	require.NoError(t, err)
	require.Equal(t, apps.Form{
		Source: &apps.Call{
			Path: "/sourcepath",
		},
		Title:  "form title",
		Header: "form header",
		Footer: "form footer",
		Icon:   "/iconpath",
		Submit: &apps.Call{
			Path: "/submitpath",
		},
		SubmitButtons: "fieldName",
		Fields: []apps.Field{
			{
				Name: "field1",
				Type: "text",
			},
			{
				Name: "field2",
				Type: "text",
			},
		},
	}, f)

	const short = `"/test"`
	f = apps.Form{}
	err = json.Unmarshal([]byte(short), &f)
	require.NoError(t, err)
	require.Equal(t, apps.Form{Source: &apps.Call{Path: "/test"}}, f)
}
