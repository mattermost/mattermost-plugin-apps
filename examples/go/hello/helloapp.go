package hello

import (
	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-plugin-apps/utils"
)

const (
	fieldUserID   = "userID"
	fieldMessage  = "message"
	fieldResponse = "response"
)

const (
	PathInstall                  = "/install"
	PathSendSurvey               = "/send"
	PathSendSurveyModal          = "/send-modal"
	PathSendSurveyCommandToModal = "/send-command-modal"
	PathSubscribeChannel         = "/subscribe"
	PathUnsubscribeChannel       = "/unsubscribe"
	PathSurvey                   = "/survey"
	PathUserJoinedChannel        = "/user-joined-channel"
	PathSubmitSurvey             = "/survey-submit"
)

type HelloApp struct {
	mm  *pluginapi.Client
	log utils.Logger
}

func NewHelloApp(mm *pluginapi.Client, log utils.Logger) *HelloApp {
	return &HelloApp{
		mm:  mm,
		log: log,
	}
}
