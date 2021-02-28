package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
)

func (h *HelloApp) UserJoinedChannel(call *apps.Call) {
	go func() {
		bot := mmclient.AsBot(call.Context)

		err := sendSurvey(bot, call.Context.UserID, "welcome to channel")
		if err != nil {
			h.mm.Log.Error("error sending survey", "err", err.Error())
		}
	}()
}
