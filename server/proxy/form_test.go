package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestCleanForm(t *testing.T) {
	type TC = struct {
		name             string
		in               apps.Form
		expectedOut      apps.Form
		expectedProblems string
	}
	testCases := []TC{
		{
			name: "no field filter on names",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Name: "field1",
					},
					{
						Name: "field2",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Name: "field1",
					},
					{
						Name: "field2",
					},
				},
			},
		},
		{
			name: "no field filter on labels",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Label: "field1",
						Name:  "same",
					},
					{
						Label: "field2",
						Name:  "same",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Label: "field1",
						Name:  "same",
					},
					{
						Label: "field2",
						Name:  "same",
					},
				},
			},
		},
		{
			name: "field filter with no name",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Label: "field1",
					},
					{
						Label: "field2",
						Name:  "same",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Label: "field2",
						Name:  "same",
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* field with no name, label field1\n\n",
		},
		{
			name: "field filter with same label inferred from name",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeBool,
						Name: "same",
					},
					{
						Type: apps.FieldTypeChannel,
						Name: "same",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeBool,
						Name: "same",
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated label: \"same\" (field: same)\n\n",
		},
		{
			name: "field filter with same label",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
					{
						Type:  apps.FieldTypeChannel,
						Label: "same",
						Name:  "field2",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated label: \"same\" (field: field2)\n\n",
		},
		{
			name: "field filter with same label",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
					{
						Type:  apps.FieldTypeChannel,
						Label: "same",
						Name:  "field2",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated label: \"same\" (field: field2)\n\n",
		},
		{
			name: "field filter with multiword name",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "multiple word",
						Name:  "multiple word",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{},
			},
			expectedProblems: "1 error occurred:\n\t* field name must be a single word: \"multiple word\"\n\n",
		},
		{
			name: "field filter with multiword label",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "multiple word",
						Name:  "singleword",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{},
			},
			expectedProblems: "1 error occurred:\n\t* label must be a single word: \"multiple word\" (field: singleword)\n\n",
		},
		{
			name: "field filter more than one field",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
					{
						Type:  apps.FieldTypeChannel,
						Label: "same",
						Name:  "field2",
					},
					{
						Type:  apps.FieldTypeText,
						Label: "same",
						Name:  "field3",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type:  apps.FieldTypeBool,
						Label: "same",
						Name:  "field1",
					},
				},
			},
			expectedProblems: "2 errors occurred:\n\t* repeated label: \"same\" (field: field2)\n\t* repeated label: \"same\" (field: field3)\n\n",
		},
		{
			name: "field filter static with no options",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{},
			},
			expectedProblems: "1 error occurred:\n\t* no options for static select: field1\n\n",
		},
		{
			name: "field filter static options with no label",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Value: "opt1",
							},
							{},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Value: "opt1",
							},
						},
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* option with neither label nor value (field field1)\n\n",
		},
		{
			name: "field filter static options with same label inferred from value",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Value:    "same",
								IconData: "opt1",
							},
							{
								Value:    "same",
								IconData: "opt2",
							},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Value:    "same",
								IconData: "opt1",
							},
						},
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated label \"same\" on select option (field field1)\n\n",
		},
		{
			name: "field filter static options with same label",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "same",
								Value: "opt1",
							},
							{
								Label: "same",
								Value: "opt2",
							},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "same",
								Value: "opt1",
							},
						},
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated label \"same\" on select option (field field1)\n\n",
		},
		{
			name: "field filter static options with same value",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "opt1",
								Value: "same",
							},
							{
								Label: "opt2",
								Value: "same",
							},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "opt1",
								Value: "same",
							},
						},
					},
				},
			},
			expectedProblems: "1 error occurred:\n\t* repeated value \"same\" on select option (field field1)\n\n",
		},
		{
			name: "invalid static options don't consume namespace",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "same1",
								Value: "same1",
							},
							{
								Label: "same1",
								Value: "same2",
							},
							{
								Label: "same2",
								Value: "same1",
							},
							{
								Label: "same2",
								Value: "same2",
							},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{
								Label: "same1",
								Value: "same1",
							},
							{
								Label: "same2",
								Value: "same2",
							},
						},
					},
				},
			},
			expectedProblems: "2 errors occurred:\n\t* repeated label \"same1\" on select option (field field1)\n\t* repeated value \"same1\" on select option (field field1)\n\n",
		},
		{
			name: "field filter static with no valid options",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{},
						},
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{},
			},
			expectedProblems: "2 errors occurred:\n\t* option with neither label nor value (field field1)\n\t* no options for static select: field1\n\n",
		},
		{
			name: "invalid static field does not consume namespace",
			in: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Type: apps.FieldTypeStaticSelect,
						Name: "field1",
						SelectStaticOptions: []apps.SelectOption{
							{},
						},
					},
					{
						Name: "field1",
					},
				},
			},
			expectedOut: apps.Form{
				Title:  "Test",
				Submit: apps.NewCall("/url"),
				Fields: []apps.Field{
					{
						Name: "field1",
					},
				},
			},
			expectedProblems: "2 errors occurred:\n\t* option with neither label nor value (field field1)\n\t* no options for static select: field1\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := cleanForm(tc.in)

			require.Equal(t, tc.expectedOut, out)
			if tc.expectedProblems != "" {
				require.Error(t, err)
				require.Equal(t, tc.expectedProblems, err.Error())
			}
		})
	}
}
