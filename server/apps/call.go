package apps

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/store"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type Call struct {
	FormURL string        `json:"form_url,omitempty"`
	Values  FormValues    `json:"values,omitempty"`
	Context *Context      `json:"context,omitempty"`
	Expand  *store.Expand `json:"expand,omitempty"`
	AsModal bool          `json:"as_modal,omitempty"`
	From    []*Location   `json:"from,omitempty"`
}

type CallResponseType string

const (
	CallResponseTypeCall     = CallResponseType("call")
	CallResponseTypeOK       = CallResponseType("ok")
	CallResponseTypeNavigate = CallResponseType("navigate")
	CallResponseTypeError    = CallResponseType("error")
	CallResponseTypeCommand  = CallResponseType("command")
)

type CallResponse struct {
	Type CallResponseType `json:"type,omitempty"`

	Markdown md.MD                  `json:"markdown,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`

	Error string `json:"error,omitempty"`

	URL                string `json:"url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	Call *Call `json:"call,omitempty"`
}

type FormValues struct {
	Data map[string]interface{} `json:"data"`
	Raw  string                 `json:"raw"`
}

func (s *service) Call(call *Call) (*CallResponse, error) {
	var err error
	req := *call
	// TODO Expand using the App's bot credentials!
	req.Context, err = s.newExpander(call.Context).Expand(call.Expand)
	if err != nil {
		return nil, err
	}
	req.Expand = nil
	req.FormURL = ""

	return s.Client.PostCall(call)
}

func (fv *FormValues) Get(name string) string {
	if fv == nil || fv.Data == nil {
		return ""
	}
	value, ok := fv.Data[name].(string)
	if !ok {
		return ""
	}

	return value
}

func UnmarshalCallData(data []byte) (*Call, error) {
	call := Call{
		Context: &Context{},
	}
	err := json.Unmarshal(data, &call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}

func UnmarshalCallReader(in io.Reader) (*Call, error) {
	call := Call{
		Context: &Context{},
	}
	err := json.NewDecoder(in).Decode(&call)
	if err != nil {
		return nil, err
	}
	return &call, nil
}
