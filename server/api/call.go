package api

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type CallType string

const (
	CallTypeSubmit = CallType("")
	CallTypeCancel = CallType("cancel")
	CallTypeForm   = CallType("form")
)

type Call struct {
	URL        string            `json:"url,omitempty"`
	Type       CallType          `json:"type,omitempty"`
	Values     map[string]string `json:"values,omitempty"`
	Context    *Context          `json:"context,omitempty"`
	AsModal    bool              `json:"as_modal,omitempty"`
	RawCommand string            `json:"raw_command,omitempty"`
	Expand     *Expand           `json:"expand,omitempty"`
}

type CallResponseType string

// TODO <><> ticket: Call and Command should be scoped and retricted, TBD
const (
	CallResponseTypeOK        = CallResponseType("")
	CallResponseTypeError     = CallResponseType("error")
	CallResponseTypeForm      = CallResponseType("form")
	CallResponseTypeCall      = CallResponseType("call")
	CallResponseTypeCommand   = CallResponseType("command")
	CallResponseTypeNavigate  = CallResponseType("navigate")
	CallResponseTypeOpenModal = CallResponseType("open_modal")
)

type CallResponse struct {
	Type CallResponseType `json:"type,omitempty"`

	Markdown md.MD                  `json:"markdown,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`

	Error string `json:"error,omitempty"`

	NavigateToURL      string `json:"navigate_to_url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	Call *Call `json:"call,omitempty"`
	Form *Form `json:"form,omitempty"`
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

func MakeCall(url string, namevalues ...string) *Call {
	c := &Call{
		URL: url,
	}

	values := map[string]string{}
	for len(namevalues) > 0 {
		switch len(namevalues) {
		case 1:
			values[namevalues[0]] = ""
			namevalues = namevalues[1:]

		default:
			values[namevalues[0]] = namevalues[1]
			namevalues = namevalues[2:]
		}
	}
	if len(values) > 0 {
		c.Values = values
	}
	return c
}
