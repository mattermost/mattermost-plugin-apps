package apps

import (
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Wish struct {
	URL string
}

type Call struct {
	// Only one of Wish or Modal can be set
	Wish  *Wish  `json:"wish,omitempty"`
	Modal *Modal `json:"modal,omitempty"`

	Data *CallData `json:"data"`
}

type CallData struct {
	Context  CallContext `json:"context"`
	Values   FormValues  `json:"values,omitempty"`
	Expanded *Expanded   `json:"expanded,omitempty"`
	From     []*Location `json:"from,omitempty"`
}

type CallContext struct {
	// For convenience, to use in go-land to pass the AppID around
	AppID AppID `json:"-"`

	// ActingUserID is the Mattermost User ID of the acting user
	ActingUserID string `json:"acting_user_id"`

	// TeamID, ChannelID, PostID represent the "location" in Mattermost that the
	// call is associated with. TeamID is usually set, ChannelID and PostID are
	// optional.
	TeamID    string `json:"team_id"`
	ChannelID string `json:"channel_id,omitempty"`
	PostID    string `json:"post_id,omitempty"`

	LogTo *Thread `json:"log_to,omitempty"`

	Props map[string]string `json:"props,omitempty"`
}

type Thread struct {
	ChannelID  string `json:"channel_id"`
	RootPostID string `json:"root_post_id"`
}

func (c *CallContext) Get(n string) string {
	if len(c.Props) == 0 {
		return ""
	}
	return c.Props[n]
}

func (c *CallContext) Set(n, v string) {
	if len(c.Props) == 0 {
		c.Props = map[string]string{}
	}
	c.Props[n] = v
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
