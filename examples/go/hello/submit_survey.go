package hello

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-server/v5/model"
)

func SubmitSurvey(c *apps.CallRequest) *apps.CallResponse {
	location := strings.Split(string(c.Context.Location), "/")
	if len(location) == 0 {
		return &apps.CallResponse{
			Type:      apps.CallResponseTypeError,
			ErrorText: "Wrong location.",
		}
	}
	selected := location[len(location)-1]
	if selected == "button" {
		bot := mmclient.AsBot(c.Context)
		p := &model.Post{
			Id:      c.Context.PostID,
			Message: "The survey will not be sent",
		}
		_, _ = bot.UpdatePost(c.Context.PostID, p)
	}
	return &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: fmt.Sprintf("You answered the survey with `%s`.", selected),
	}
}
