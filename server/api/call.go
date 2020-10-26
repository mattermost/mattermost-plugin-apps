package api

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Call struct {
	FormURL string      `json:"form_url,omitempty"`
	Values  FormValues  `json:"values,omitempty"`
	Context *Context    `json:"context,omitempty"`
	Expand  *Expand     `json:"expand,omitempty"`
	AsModal bool        `json:"as_modal,omitempty"`
	From    []*Location `json:"from,omitempty"`
}

type CallResponseType string

const (
	CallResponseTypeCall     = CallResponseType("call")
	CallResponseTypeOK       = CallResponseType("ok")
	CallResponseTypeNavigate = CallResponseType("navigate")
	CallResponseTypeError    = CallResponseType("error")
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

type FormValues struct {
	Data map[string]interface{} `json:"data"`
	Raw  string                 `json:"raw"`
}

func (fv *FormValues) Get(name string) string {
	if fv == nil || fv.Data == nil {
		return ""
	}
	return fv.Data[name].(string)
}

func UnmarshalCallFromData(data []byte) (*Call, error) {
	call := Call{
		Context: &Context{},
	}
	err := json.Unmarshal(data, &call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func UnmarshalCallFromReader(in io.Reader) (*Call, error) {
	call := Call{
		Context: &Context{},
	}
	err := json.NewDecoder(in).Decode(&call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}
