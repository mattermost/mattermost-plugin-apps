package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/modelapps"
)

func (h *HelloApp) UserJoinedChannel(call *modelapps.Call) {
	go func() {
		bot := modelapps.AsBot(call.Context)

		err := sendSurvey(bot, call.Context.UserID, "welcome to channel")
		if err != nil {
			h.API.Mattermost.Log.Error("error sending survey", "err", err.Error())
		}
	}()
}
