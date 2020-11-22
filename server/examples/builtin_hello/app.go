package builtin_hello

import (
	"github.com/mattermost/mattermost-plugin-apps/server/api"
	"github.com/pkg/errors"
)

const (
	AppID          = "builtin"
	AppDisplayName = "builtin hello display name"
	AppDescription = "builtin hello description"
)

const (
	fieldUserID   = "userID"
	fieldMessage  = "message"
	fieldResponse = "response"
)

const (
	PathSendSurvey       = "/send"
	PathSubscribeChannel = "/subscribe"
	PathInstall          = "/install"
	PathSurvey           = "/survey"
)

type App struct{}

var _ api.Upstream = (*App)(nil)

var calls = map[string]func(*api.Call) *api.CallResponse{
	PathInstall:    Install,
	PathSendSurvey: SendSurvey,
	PathSurvey:     Survey,
}

func (a *App) InvokeCall(c *api.Call) (*api.CallResponse, error) {
	f := calls[c.URL]
	if f == nil {
		return nil, errors.Errorf("%s is not found", c.URL)
	}
	return f(c), nil
}

func (a *App) InvokeNotification(n *api.Notification) error {
	if n.Subject != api.SubjectUserJoinedChannel {
		return errors.Errorf("%s is supported", n.Subject)
	}
	NotifyUserJoinedChannel(n)
	return nil
}

func callError(err error) *api.CallResponse {
	return &api.CallResponse{
		Type:  api.CallResponseTypeError,
		Error: err.Error(),
	}
}
