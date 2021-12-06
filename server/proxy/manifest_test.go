package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func TestMergeDeployData(t *testing.T) {
	aws := apps.AWSLambda{
		Functions: []apps.AWSLambdaFunction{
			{},
		},
	}
	newAWS := apps.AWSLambda{
		Functions: []apps.AWSLambdaFunction{
			{},
			{},
		},
	}

	openFAAS := apps.OpenFAAS{
		Functions: []apps.OpenFAASFunction{
			{},
		},
	}

	newOpenFAAS := apps.OpenFAAS{
		Functions: []apps.OpenFAASFunction{
			{},
			{},
		},
	}

	http := apps.HTTP{
		RootURL: "test",
	}
	newHTTP := apps.HTTP{
		RootURL: "test1",
	}

	plugin := apps.Plugin{
		PluginID: "test",
	}
	newPlugin := apps.Plugin{
		PluginID: "test1",
	}

	for _, tc := range []struct {
		name     string
		prevd    apps.Deploy
		newd     apps.Deploy
		add      apps.DeployTypes
		remove   apps.DeployTypes
		expected apps.Deploy
	}{
		{
			name: "same upstreams, no add-remove",
			prevd: apps.Deploy{
				AWSLambda: &aws,
				HTTP:      &http,
				OpenFAAS:  &openFAAS,
				Plugin:    &plugin,
			},
			newd: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				OpenFAAS:  &newOpenFAAS,
				Plugin:    &newPlugin,
			},
			expected: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				OpenFAAS:  &newOpenFAAS,
				Plugin:    &newPlugin,
			},
		},
		{
			name: "new is subset of old, no add-remove",
			prevd: apps.Deploy{
				HTTP:   &http,
				Plugin: &plugin,
			},
			newd: apps.Deploy{
				Plugin: &newPlugin,
			},
			expected: apps.Deploy{
				HTTP:   &http,
				Plugin: &newPlugin,
			},
		},
		{
			name: "old is subset of new, no add-remove",
			prevd: apps.Deploy{
				Plugin: &plugin,
			},
			newd: apps.Deploy{
				HTTP:   &newHTTP,
				Plugin: &newPlugin,
			},
			expected: apps.Deploy{
				Plugin: &newPlugin,
			},
		},
		{
			name: "old is subset of new, no add-remove",
			prevd: apps.Deploy{
				Plugin: &plugin,
			},
			newd: apps.Deploy{
				HTTP:   &newHTTP,
				Plugin: &newPlugin,
			},
			expected: apps.Deploy{
				Plugin: &newPlugin,
			},
		},
		{
			name:  "old is empty, no add-remove",
			prevd: apps.Deploy{},
			newd: apps.Deploy{
				HTTP:   &newHTTP,
				Plugin: &newPlugin,
			},
			expected: apps.Deploy{},
		},
		{
			name: "new is empty, no add-remove",
			prevd: apps.Deploy{
				HTTP:   &http,
				Plugin: &plugin,
			},
			newd: apps.Deploy{},
			expected: apps.Deploy{
				HTTP:   &http,
				Plugin: &plugin,
			},
		},
		{
			name: "everything",
			prevd: apps.Deploy{
				OpenFAAS: &openFAAS,
				Plugin:   &plugin,
			},
			newd: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				OpenFAAS:  &newOpenFAAS,
			},
			add:    []apps.DeployType{apps.DeployAWSLambda},
			remove: []apps.DeployType{apps.DeployOpenFAAS},
			expected: apps.Deploy{
				AWSLambda: &newAWS,
				Plugin:    &plugin,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := mergeDeployData(tc.prevd, tc.newd, tc.add, tc.remove)
			require.EqualValues(t, tc.expected, d)
		})
	}
}
