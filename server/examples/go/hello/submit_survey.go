package hello

import (
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

func SubmitSurvey(c *api.Call) *api.CallResponse {
	location := strings.Split(string(c.Context.Location), "/")
	if len(location) == 0 {
		return &api.CallResponse{
			Type:      api.CallResponseTypeError,
			ErrorText: "Wrong location.",
		}
	}
	selected := location[len(location)-1]
	return &api.CallResponse{
		Type:     api.CallResponseTypeOK,
		Markdown: md.Markdownf("You answered the survey with `%s`.", selected),
	}
}
