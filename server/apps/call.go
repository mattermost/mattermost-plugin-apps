package apps

import "github.com/mattermost/mattermost-plugin-apps/server/utils/md"

type Params struct {
	Values   FormValues  `json:"values,omitempty"`
	Expanded *Expanded   `json:"expanded,omitempty"`
	Env      FormValues  `json:"env,omitempty"`
	From     []*Location `json:"from,omitempty"`
}

type WishCall struct {
	Wish *Wish
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
	callOKProps
	callErrorProps
	callGotoProps
}

type callOKProps struct {
	Markdown md.MD
	Data     map[string]interface{}
}

type callErrorProps struct {
	Error error
}

type callGotoProps struct {
	URL                string
	UseExternalBrowser bool
}

type callCall struct {
	CallRequest *CallRequest
}
