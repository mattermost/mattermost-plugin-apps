// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestDecodeField(t *testing.T) {
	var normalizedJSON = `
	{
		"name": "fieldName",
		"type": "dynamic_select",
		"is_required": true,
		"readonly": true,
		"value": "something",
		"description": "some description",
		"label": "fieldLabel",
		"hint": "something hint",
		"position": 1,
		"modal_label": "fieldModalLabel",
		"multiselect": true,
		"refresh": true,
		"options": [
			{
				"label": "l1",
				"value": "v1"
			},
			{
				"label": "l2",
				"value": "v2"
			}
		],
		"lookup": {
			"path": "/fieldLookupPath",
			"expand": {
				"acting_user": "all"
			}
		}
	}
	`

	var shortJSON = `
	{
		"name": "fieldName",
		"type": "dynamic_select",
		"is_required": true,
		"readonly": true,
		"value": "something",
		"description": "some description",
		"label": "fieldLabel",
		"hint": "something hint",
		"position": 1,
		"modal_label": "fieldModalLabel",
		"multiselect": true,
		"refresh": true,
		"options": [
			{
				"label": "l1",
				"value": "v1"
			},
			{
				"label": "l2",
				"value": "v2"
			}
		],
		"lookup": {
			"path": "/fieldLookupPath"
		}
	}
	`

	for name, tc := range map[string]struct {
		in            string
		expected      apps.Field
		expectedError bool
	}{
		"normalized": {
			in: normalizedJSON,
			expected: apps.Field{
				Name:                 "fieldName",
				Type:                 "dynamic_select",
				IsRequired:           true,
				ReadOnly:             true,
				Value:                "something",
				Description:          "some description",
				Label:                "fieldLabel",
				AutocompleteHint:     "something hint",
				AutocompletePosition: 1,
				ModalLabel:           "fieldModalLabel",
				SelectIsMulti:        true,
				SelectRefresh:        true,
				StaticSelectOptions: []apps.SelectOption{
					{
						Label: "l1",
						Value: "v1",
					},
					{
						Label: "l2",
						Value: "v2",
					},
				},
				DynamicSelectLookup: &apps.Call{
					Path: "/fieldLookupPath",
					Expand: &apps.Expand{
						ActingUser: apps.ExpandAll,
					},
				},
			},
		},
		"short": {
			in: shortJSON,
			expected: apps.Field{
				Name:                 "fieldName",
				Type:                 "dynamic_select",
				IsRequired:           true,
				ReadOnly:             true,
				Value:                "something",
				Description:          "some description",
				Label:                "fieldLabel",
				AutocompleteHint:     "something hint",
				AutocompletePosition: 1,
				ModalLabel:           "fieldModalLabel",
				SelectIsMulti:        true,
				SelectRefresh:        true,
				StaticSelectOptions: []apps.SelectOption{
					{
						Label: "l1",
						Value: "v1",
					},
					{
						Label: "l2",
						Value: "v2",
					},
				},
				DynamicSelectLookup: apps.NewCall("/fieldLookupPath"),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := apps.Field{}
			err := json.Unmarshal([]byte(tc.in), &f)

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, f)
			}
		})
	}
}
