package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildConfig(t *testing.T) {
	assert.NotEmpty(t, BuildDate)
	assert.NotEmpty(t, BuildHash)
	assert.NotEmpty(t, BuildHashShort)
}
