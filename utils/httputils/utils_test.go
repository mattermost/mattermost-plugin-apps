// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package httputils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeRemoteBaseURL(t *testing.T) {
	for _, tc := range []struct {
		in, siteURL, out, err string
	}{
		// Happy
		{"http://mmtest.somedomain.net", "", "http://mmtest.somedomain.net", ""},
		{"https://mmtest.somedomain.net", "", "https://mmtest.somedomain.net", ""},
		{"some://mmtest.somedomain.net", "", "some://mmtest.somedomain.net", ""},
		{"mmtest.somedomain.net", "", "https://mmtest.somedomain.net", ""},
		{"mmtest.somedomain.net/", "", "https://mmtest.somedomain.net", ""},
		{"mmtest.somedomain.net/abc", "", "https://mmtest.somedomain.net/abc", ""},
		{"mmtest.somedomain.net/abc/", "", "https://mmtest.somedomain.net/abc", ""},
		{"mmtest", "", "https://mmtest", ""},
		{"mmtest/", "", "https://mmtest", ""},
		{"//xyz.com", "", "https://xyz.com", ""},
		{"//xyz.com/", "", "https://xyz.com", ""},

		// Errors
		{"[jdsh", "", "",
			`parse "//[jdsh": missing ']' in host`},
		{"/mmtest", "", "",
			`invalid URL, no hostname: "/mmtest"`},
		{"/mmtest/", "", "",
			`invalid URL, no hostname: "/mmtest/"`},
		{"http:/mmtest/", "", "",
			`invalid URL, no hostname: "http:/mmtest/"`},
		{"hƒƒp://xyz.com", "", "",
			`parse "hƒƒp://xyz.com": first path segment in URL cannot contain colon`},
		{"https://mattermost.site.url", "https://mattermost.site.url/", "",
			"https://mattermost.site.url is the Mattermost site URL. Please use the remote application's URL"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, err := NormalizeRemoteBaseURL(tc.siteURL, tc.in)
			require.Equal(t, tc.out, out)
			errTxt := ""
			if err != nil {
				errTxt = err.Error()
			}
			require.Equal(t, tc.err, errTxt)
		})
	}
}
