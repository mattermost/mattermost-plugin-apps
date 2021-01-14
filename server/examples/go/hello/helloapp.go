package hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
)

const (
	fieldUserID   = "userID"
	fieldMessage  = "message"
	fieldResponse = "response"
)

const (
	PathSendSurvey        = "/send"
	PathSubscribeChannel  = "/subscribe"
	PathSurvey            = "/survey"
	PathUserJoinedChannel = "/user-joined-channel"
	PathEmbedSurveySubmit = "/embedded"
)

type HelloApp struct {
	API *api.Service
}

func NewHelloApp(api *api.Service) *HelloApp {
	return &HelloApp{api}
}
