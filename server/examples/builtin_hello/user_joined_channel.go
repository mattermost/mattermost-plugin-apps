package builtin_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
)

func NotifyUserJoinedChannel(n *api.Notification) {
	go func() {
		bot := examples.AsBot(n.Context)
		err := sendSurvey(bot, n.Context.UserID, "welcome to channel")
		if err != nil {
			// h.api.Mattermost.Log.Error("error sending survey", "err", err.Error())
		}
	}()
}
