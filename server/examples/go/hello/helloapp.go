package hello

import pluginapi "github.com/mattermost/mattermost-plugin-api"

const (
	fieldUserID   = "userID"
	fieldMessage  = "message"
	fieldResponse = "response"
)

const (
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
	mm *pluginapi.Client
}

func NewHelloApp(mm *pluginapi.Client) *HelloApp {
	return &HelloApp{
		mm: mm,
	}
}
