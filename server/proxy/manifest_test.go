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

	kubeless := apps.Kubeless{
		Functions: []apps.KubelessFunction{
			{},
		},
	}
	newKubeless := apps.Kubeless{
		Functions: []apps.KubelessFunction{
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
				Kubeless:  &kubeless,
				OpenFAAS:  &openFAAS,
				Plugin:    &plugin,
			},
			newd: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				Kubeless:  &newKubeless,
				OpenFAAS:  &newOpenFAAS,
				Plugin:    &newPlugin,
			},
			expected: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				Kubeless:  &newKubeless,
				OpenFAAS:  &newOpenFAAS,
				Plugin:    &newPlugin,
			},
		},
		{
			name: "new is subset of old, no add-remove",
			prevd: apps.Deploy{
				HTTP:     &http,
				Plugin:   &plugin,
				Kubeless: &kubeless,
			},
			newd: apps.Deploy{
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
			expected: apps.Deploy{
				HTTP:     &http,
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
		},
		{
			name: "old is subset of new, no add-remove",
			prevd: apps.Deploy{
				Plugin:   &plugin,
				Kubeless: &kubeless,
			},
			newd: apps.Deploy{
				HTTP:     &newHTTP,
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
			expected: apps.Deploy{
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
		},
		{
			name: "old is subset of new, no add-remove",
			prevd: apps.Deploy{
				Plugin:   &plugin,
				Kubeless: &kubeless,
			},
			newd: apps.Deploy{
				HTTP:     &newHTTP,
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
			expected: apps.Deploy{
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
		},
		{
			name:  "old is empty, no add-remove",
			prevd: apps.Deploy{},
			newd: apps.Deploy{
				HTTP:     &newHTTP,
				Plugin:   &newPlugin,
				Kubeless: &newKubeless,
			},
			expected: apps.Deploy{},
		},
		{
			name: "new is empty, no add-remove",
			prevd: apps.Deploy{
				HTTP:     &http,
				Plugin:   &plugin,
				Kubeless: &kubeless,
			},
			newd: apps.Deploy{},
			expected: apps.Deploy{
				HTTP:     &http,
				Plugin:   &plugin,
				Kubeless: &kubeless,
			},
		},
		{
			name: "everything",
			prevd: apps.Deploy{
				Kubeless: &kubeless,
				OpenFAAS: &openFAAS,
				Plugin:   &plugin,
			},
			newd: apps.Deploy{
				AWSLambda: &newAWS,
				HTTP:      &newHTTP,
				Kubeless:  &newKubeless,
				OpenFAAS:  &newOpenFAAS,
			},
			add:    []apps.DeployType{apps.DeployAWSLambda},
			remove: []apps.DeployType{apps.DeployKubeless, apps.DeployOpenFAAS},
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
