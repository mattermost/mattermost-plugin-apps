package hello

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-server/v5/model"
)

func SubmitSurvey(c *apps.Call) *apps.CallResponse {
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
		_, err := bot.UpdatePost(c.Context.PostID, p)
		fmt.Println(err)
		fmt.Println(c.Context.PostID)
	}
	return &apps.CallResponse{
		Type:     apps.CallResponseTypeOK,
		Markdown: md.Markdownf("You answered the survey with `%s`.", selected),
	}
}
