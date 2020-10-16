package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
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

	Error string `json:"error,omitempty"`

	URL                string `json:"url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	Call *Call `json:"call,omitempty"`
}
