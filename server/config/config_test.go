package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildConfig(t *testing.T) {
	assert.NotEmpty(t, BuildDate)
	assert.NotEmpty(t, BuildHash)
	assert.NotEmpty(t, BuildHashShort)
}

func TestAllowHTTPRegexp(t *testing.T) {
	require.False(t, allowHTTPAppsDomains.MatchString("dkh-apps-upgrade.cloud.mattermost.com"))
	require.True(t, allowHTTPAppsDomains.MatchString("dkh-apps-upgrade.test.mattermost.cloud"))
	require.True(t, allowHTTPAppsDomains.MatchString("community.mattermost.com"))
	require.True(t, allowHTTPAppsDomains.MatchString("community-daily.mattermost.com"))
	require.True(t, allowHTTPAppsDomains.MatchString("community-release.mattermost.com"))
}
