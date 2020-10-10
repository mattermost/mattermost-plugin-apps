package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/pkg/errors"
)

type Call struct {
	// TODO consider identifying Call target by (meta) URL
	// Only one of Wish or Modal can be set
	Wish  *store.Wish `json:"wish,omitempty"`
	Modal *Modal      `json:"modal,omitempty"`

	Request *CallRequest `json:"data"`
}

type CallRequest struct {
	Context *Context    `json:"context"`
	From    []*Location `json:"from,omitempty"`
	Values  FormValues  `json:"values,omitempty"`
}

func NewCallRequest(cc *Context) *CallRequest {
	clone := Context{}
	if cc != nil {
		clone = *cc
	}
	if clone.expandedContext == nil {
		clone.expandedContext = &expandedContext{}
	}

	return &CallRequest{
		Context: &clone,
		Values: FormValues{
			Data: map[string]interface{}{},
		},
	}
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

	Markdown md.MD       `json:"markdown,omitempty"`
	Data     interface{} `json:"data,omitempty"`

	Error error `json:"error,omitempty"`

	URL                string `json:"url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	Call *Call `json:"call,omitempty"`
}

func (s *Service) Call(call *Call) (*CallResponse, error) {
	switch {
	case call.Wish != nil && call.Modal == nil:
		return s.PostWish(call)
	case call.Modal != nil && call.Wish == nil:
		return s.CallModal(call)
	default:
		return nil, errors.New("invalid Call, only one of Wish, Modal can be specified")
	}
}
