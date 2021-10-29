package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	assert.NotEmpty(t, BuildDate)
	assert.NotEmpty(t, BuildHash)
	assert.NotEmpty(t, BuildHashShort)
}
