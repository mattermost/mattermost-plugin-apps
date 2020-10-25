package api

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Call struct {
	URL        string            `json:"url,omitempty"`
	Values     map[string]string `json:"values,omitempty"`
	Context    *Context          `json:"context,omitempty"`
	AsModal    bool              `json:"as_modal,omitempty"`
	RawCommand string            `json:"raw_command,omitempty"`
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

func UnmarshalCallFromData(data []byte) (*Call, error) {
	call := Call{}
	err := json.Unmarshal(data, &call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func UnmarshalCallFromReader(in io.Reader) (*Call, error) {
	call := Call{}
	err := json.NewDecoder(in).Decode(&call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}
