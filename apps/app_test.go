package apps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppIDIsValid(t *testing.T) {
	t.Parallel()

	for id, valid := range map[string]bool{
		"":                                  false,
		"a":                                 false,
		"ab":                                false,
		"abc":                               true,
		"abcdefghijklmnopqrstuvwxyzabcdef":  true,
		"abcdefghijklmnopqrstuvwxyzabcdefg": false,
		"../path":                           false,
		"/etc/passwd":                       false,
		"com.mattermost.app-0.9":            true,
		"CAPS-ARE-FINE":                     true,
		"....DOTS.ALSO.......":              true,
		"----SLASHES-ALSO----":              true,
		"___AND_UNDERSCORES____":            true,
	} {
		t.Run(id, func(t *testing.T) {
			err := AppID(id).IsValid()
			if valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
