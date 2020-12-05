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
	PathBindings          = api.AppBindingsPath // convention for Mattermost Apps
	PathInstall           = api.AppInstallPath  // convention for Mattermost Apps
	PathSendSurvey        = "/send"
	PathSubscribeChannel  = "/subscribe"
	PathSurvey            = "/survey"
	PathUserJoinedChannel = "/user-joined-channel"
)

type HelloApp struct {
	API *api.Service
}

func NewHelloApp(api *api.Service) *HelloApp {
	return &HelloApp{api}
}
