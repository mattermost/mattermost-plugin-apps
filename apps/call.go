package apps

import (
	"encoding/json"
	"io"

	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
)

type CallType string

const (
	// CallTypeSubmit (default) indicates the intent to take action.
	CallTypeSubmit = CallType("submit")
	// CallTypeForm retrieves the form definition for the current set of values,
	// and the context.
	CallTypeForm = CallType("form")
	// CallTypeCancel is used for for the (rare?) case of when the form with
	// SubmitOnCancel set is dismissed by the user.
	CallTypeCancel = CallType("cancel")
	// CallTypeLookup is used to fetch items for dynamic select elements
	CallTypeLookup = CallType("lookup")
)

// A Call invocation is supplied a BotAccessToken as part of the context. If a
// call needs acting user's or admin tokens, it should be specified in the
// Expand section.
//
// If a user or admin token are required and are not available from previous
// consent, the appropriate OAuth flow is launched, and the Call is executed
// upon its success.
//
// TODO: what if a call needs a token and it was not provided? Return a call to
// itself with Expand.
type Call struct {
	Path   string      `json:"path,omitempty"`
	Expand *Expand     `json:"expand,omitempty"`
	State  interface{} `json:"state,omitempty"`
}

type CallRequest struct {
	Call
	Type          CallType               `json:"type"`
	Values        map[string]interface{} `json:"values,omitempty"`
	Context       *Context               `json:"context,omitempty"`
	RawCommand    string                 `json:"raw_command,omitempty"`
	SelectedField string                 `json:"selected_field,omitempty"`
	Query         string                 `json:"query,omitempty"`
}

type CallResponseType string

const (
	// CallResponseTypeOK indicates that the call succeeded, and returns
	// Markdown and Data.
	CallResponseTypeOK = CallResponseType("ok")

	// CallResponseTypeOK indicates an error, returns Error.
	CallResponseTypeError = CallResponseType("error")

	// CallResponseTypeForm returns the definition of the form to display for
	// the inputs.
	CallResponseTypeForm = CallResponseType("form")

	// CallResponseTypeCall indicates that another Call that should be executed
	// (from the user-agent?). Call field is returned.
	CallResponseTypeCall = CallResponseType("call")

	// CallResponseTypeNavigate indicates that the user should be forcefully
	// navigated to a URL, which may be a channel in Mattermost. NavigateToURL
	// and UseExternalBrowser are expected to be returned.
	// TODO should CallResponseTypeNavigate be a variation of CallResponseTypeOK?
	CallResponseTypeNavigate = CallResponseType("navigate")
)

type CallResponse struct {
	Type CallResponseType `json:"type"`

	// Used in CallResponseTypeOK to return the displayble, and JSON results
	Markdown md.MD       `json:"markdown,omitempty"`
	Data     interface{} `json:"data,omitempty"`

	// Used in CallResponseTypeError
	ErrorText string `json:"error,omitempty"`

	// Used in CallResponseTypeNavigate
	NavigateToURL      string `json:"navigate_to_url,omitempty"`
	UseExternalBrowser bool   `json:"use_external_browser,omitempty"`

	// Used in CallResponseTypeCall
	Call *Call `json:"call,omitempty"`

	// Used in CallResponseTypeForm
	Form *Form `json:"form,omitempty"`
}

func NewErrorCallResponse(err error) *CallResponse {
	return &CallResponse{
		Type: CallResponseTypeError,
		// TODO <><> ticket use MD and Data, remove Error
		ErrorText: err.Error(),
	}
}

// Error() makes CallResponse a valid error, for convenience
func (cr *CallResponse) Error() string {
	if cr.Type == CallResponseTypeError {
		return cr.ErrorText
	}
	return ""
}

func UnmarshalCallRequestFromData(data []byte) (*CallRequest, error) {
	c := CallRequest{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func UnmarshalCallRequestFromReader(in io.Reader) (*CallRequest, error) {
	c := CallRequest{}
	err := json.NewDecoder(in).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func MakeCall(url string) *Call {
	c := &Call{
		Path: url,
	}
	return c
}

func (c *CallRequest) GetValue(name, defaultValue string) string {
	if len(c.Values) == 0 {
		return defaultValue
	}

	s, ok := c.Values[name].(string)
	if ok && s != "" {
		return s
	}

	opt, ok := c.Values[name].(map[string]interface{})
	if ok {
		if v, ok2 := opt["value"].(string); ok2 {
			return v
		}
	}

	return defaultValue
}
