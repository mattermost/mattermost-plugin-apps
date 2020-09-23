package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/pkg/errors"
)

type Call struct {
	// Only one of Wish or Modal can be set
	Wish  *Wish  `json:"wish,omitempty"`
	Modal *Modal `json:"modal,omitempty"`

	Data *CallData `json:"data"`
}

type CallData struct {
	Values   FormValues             `json:"values,omitempty"`
	Expanded *Expanded              `json:"expanded,omitempty"`
	Env      map[string]interface{} `json:"env,omitempty"`
	From     []*Location            `json:"from,omitempty"`
}

type CallResponseType string

const (
	ResponseTypeCallWish  = CallResponseType("call_wish")
	ResponseTypeCallModal = CallResponseType("call_modal")
	ResponseTypeOK        = CallResponseType("ok")
	ResponseTypeNavigate  = CallResponseType("navigate")
	ResponseTypeError     = CallResponseType("error")
)

type CallResponse struct {
	Type CallResponseType

	Markdown md.MD                  `json:"markdown,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`

	Error error `json:"error,omitempty"`

	URL                string `json:"url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	Call *Call `json:"call,omitempty"`
}

func (s *Service) Call(toAppID AppID, fromMattermostUserID string, c *Call) (*CallResponse, error) {
	switch {
	case c.Wish != nil && c.Modal == nil:
		return s.PostWish(toAppID, fromMattermostUserID, c.Wish, c.Data)
	case c.Modal != nil && c.Wish == nil:
		return s.CallModal(toAppID, c.Modal, c.Data)
	default:
		return nil, errors.New("invalid Call, only one of Wish, Modal can be specified")
	}
}
