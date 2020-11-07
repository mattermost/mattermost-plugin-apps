package api

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type CallType string

const (
	// CallTypeSubmit (default) indicates the intent to take action.
	CallTypeSubmit = CallType("")
	// CallTypeForm retrieves the form definition for the current set of falues,
	// and the context.
	CallTypeForm = CallType("form")
	// CallTypeCancel is used for for the (rare?) case of when the form with
	// SubmitOnCancel set is dismissed by the user.
	CallTypeCancel = CallType("cancel")
)

type Call struct {
	URL        string            `json:"url,omitempty"`
	Type       CallType          `json:"type,omitempty"`
	Values     map[string]string `json:"values,omitempty"`
	Context    *Context          `json:"context,omitempty"`
	RawCommand string            `json:"raw_command,omitempty"`
	Expand     *Expand           `json:"expand,omitempty"`
}

type CallResponseType string

const (
	CallResponseTypeOK       = CallResponseType("")
	CallResponseTypeError    = CallResponseType("error")
	CallResponseTypeForm     = CallResponseType("form")
	CallResponseTypeCall     = CallResponseType("call")
	CallResponseTypeNavigate = CallResponseType("navigate")
)

type CallResponse struct {
	Type CallResponseType `json:"type,omitempty"`

	// Used in CallResponseTypeOK to return the displayble, and JSON results
	Markdown md.MD                  `json:"markdown,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`

	// Used in CallResponseTypeError
	Error string `json:"error,omitempty"`

	// Used in CallResponseTypeNavigate
	NavigateToURL      string `json:"navigate_to_url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	// Used in CallResponseTypeCall
	Call *Call `json:"call,omitempty"`

	// Used in CallResponseTypeForm
	Form *Form `json:"form,omitempty"`
}

func UnmarshalCallFromData(data []byte) (*Call, error) {
	c := Call{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func UnmarshalCallFromReader(in io.Reader) (*Call, error) {
	c := Call{}
	err := json.NewDecoder(in).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
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

func (c *Call) GetValue(name, defaultValue string) string {
	if len(c.Values) == 0 || c.Values[name] == "" {
		return defaultValue
	}
	return c.Values[name]
}
