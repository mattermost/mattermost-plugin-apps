package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/mattermost/mattermost-plugin-apps/server/examples"
)

func (h *HelloApp) UserJoinedChannel(call *api.Call) {
	go func() {
		bot := examples.AsBot(call.Context)

		err := sendSurvey(bot, call.Context.UserID, "welcome to channel")
		if err != nil {
			h.API.Mattermost.Log.Error("error sending survey", "err", err.Error())
		}
	}()
}
