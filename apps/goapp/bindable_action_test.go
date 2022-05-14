package goapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMakeBindableAction(t *testing.T) {
	b := MakeBindableActionOrPanic("test îtem", nil)
	require.EqualValues(t, &BindableAction{
		bindable: bindable{
			name: "test îtem",
		},
		submit: apps.NewCall("/test-%C3%AEtem"),
	}, b)

	b = MakeBindableActionOrPanic("test îtem", nil,
		WithExpand(apps.Expand{
			ActingUser:            apps.ExpandSummary,
			ActingUserAccessToken: apps.ExpandAll,
		}),
	)

	// require.NoError(t, err)
	require.EqualValues(t,
		&BindableAction{
			bindable: bindable{
				name: "test îtem",
			},
			submit: apps.NewCall("/test-%C3%AEtem").WithExpand(apps.Expand{
				ActingUser:            apps.ExpandSummary,
				ActingUserAccessToken: apps.ExpandAll,
			}),
		}, b)
}
