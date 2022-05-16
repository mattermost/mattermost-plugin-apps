package goapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMakeBindableForm(t *testing.T) {
	testForm := apps.Form{
		Title:  "Test title",
		Header: "TODO",
		Fields: []apps.Field{
			{
				Name: "f1",
			},
		},
	}

	b := MakeBindableFormOrPanic("test îtem", testForm, nil)
	require.EqualValues(t, b,
		&BindableForm{
			BindableAction: &BindableAction{
				bindable: bindable{
					name: "test îtem",
				},
				submit: apps.NewCall("/test-%C3%AEtem"),
			},
			form: &testForm,
		})

	b = MakeBindableFormOrPanic("test îtem", testForm, nil,
		WithExpand(apps.Expand{
			ActingUser:            apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
		}),
	)

	// require.NoError(t, err)
	require.EqualValues(t, b,
		&BindableForm{
			BindableAction: &BindableAction{
				bindable: bindable{
					name: "test îtem",
				},
				submit: apps.NewCall("/test-%C3%AEtem").WithExpand(apps.Expand{
					ActingUser:            apps.ExpandSummary,
					ActingUserAccessToken: apps.ExpandAll,
				}),
			},
			form: &testForm,
		})
}
