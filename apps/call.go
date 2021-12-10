// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
	"fmt"
	"io"
)

// CallType determines what action is expected of a function.
type CallType string

const (
	// CallTypeSubmit (default) indicates the intent to take action.
	CallTypeSubmit CallType = "submit"

	// CallTypeForm retrieves the form definition for the current set of values,
	// and the context.
	CallTypeForm CallType = "form"

	// CallTypeCancel is used for for the (rare?) case of when the form with
	// SubmitOnCancel set is dismissed by the user.
	CallTypeCancel CallType = "cancel"

	// CallTypeLookup is used to fetch items for dynamic select elements
	CallTypeLookup CallType = "lookup"
)

// Call defines a way to invoke an App's function. Calls are used to fetch App's
// bindings, to process notifications, and to respond to user input from forms,
// bindings and command line.
type Call struct {
	// The path of the Call. For HTTP apps, the path is appended to the app's
	// RootURL. For AWS Lambda apps, it is mapped to the appropriate Lambda name
	// to invoke, and then passed in the call request.
	Path string `json:"path,omitempty"`

	// Expand specifies what extended data should be provided to the function in
	// each request's Context. It may be various auth tokens, configuration
	// data, or details of Mattermost entities such as the acting user, current
	// team and channel, etc.
	Expand *Expand `json:"expand,omitempty"`

	// Custom data that will be passed to the function in JSON, "as is".
	State interface{} `json:"state,omitempty"`
}

// CallRequest envelops all requests sent to Apps.
type CallRequest struct {
	// A copy of the Call struct that originated the request. Path and State are
	// of significance.
	Call

	// Values are all values entered by the user.
	Values map[string]interface{} `json:"values,omitempty"`

	// Context of execution, see the Context type for more information.
	Context Context `json:"context,omitempty"`

	// In case the request came from the command line, the raw text of the
	// command, as submitted by the user.
	RawCommand string `json:"raw_command,omitempty"`

	// SelectedField and Query are used in calls of type lookup, and calls type
	// form used to refresh the form upon user entry, to communicate what field
	// is selected, and what query string is already entered by the user for it.
	SelectedField string `json:"selected_field,omitempty"`
	Query         string `json:"query,omitempty"`
}

type CallResponseType string

const (
	// CallResponseTypeOK indicates that the call succeeded, returns optional
	// Markdown (message) and Data.
	CallResponseTypeOK CallResponseType = "ok"

	// CallResponseTypeOK indicates an error, returns ErrorText, and optional
	// field-level errors in Data.
	CallResponseTypeError CallResponseType = "error"

	// CallResponseTypeForm returns Form, the definition of the form to display.
	// If returned responding to a submit, causes the form to be displayed as a
	// modal.
	CallResponseTypeForm CallResponseType = "form"

	// CallResponseTypeCall indicates that another Call that should be executed
	// (all the way from the user-agent). Call is returned. NOT YET IMPLEMENTED.
	CallResponseTypeCall CallResponseType = "call"

	// CallResponseTypeNavigate indicates that the user should be forcefully
	// navigated to a URL, which may be a channel in Mattermost. NavigateToURL
	// and UseExternalBrowser are expected to be returned.
	CallResponseTypeNavigate CallResponseType = "navigate"
)

// CallResponse is general envelope for all Call responses.
//
// Submit requests expect ok, error, form, call, or navigate response types.
// Returning a "form" type in response to a submission from the user-agent
// triggers displaying a Modal. Returning a "call" type in response to a
// submission causes the call to be executed from the user-agent (NOT
// IMPLEMENTED YET)
//
// Form requests expect form or error.
//
// Lookup requests expect ok or error.
//
// In case of an error, the returned response type is "error", ErrorText
// contains the overall error text. Data contains optional, field-level errors.
type CallResponse struct {
	Type CallResponseType `json:"type"`

	// Used in CallResponseTypeOK to return the displayble, and JSON results
	Markdown string      `json:"markdown,omitempty"`
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

func NewErrorResponse(err error) CallResponse {
	return CallResponse{
		Type: CallResponseTypeError,
		// TODO <>/<> ticket use MD instead of ErrorText
		ErrorText: err.Error(),
	}
}

func NewDataResponse(data interface{}) CallResponse {
	return CallResponse{
		Type: CallResponseTypeOK,
		Data: data,
	}
}

func NewTextResponse(format string, args ...interface{}) CallResponse {
	return CallResponse{
		Type:     CallResponseTypeOK,
		Markdown: fmt.Sprintf(format, args...),
	}
}

func NewFormResponse(form Form) CallResponse {
	return CallResponse{
		Type: CallResponseTypeForm,
		Form: &form,
	}
}

func NewLookupResponse(opts []SelectOption) CallResponse {
	return NewDataResponse(struct {
		Items []SelectOption `json:"items"`
	}{opts})
}

// Error() makes CallResponse a valid error, for convenience
func (cr CallResponse) Error() string {
	if cr.Type == CallResponseTypeError {
		return cr.ErrorText
	}
	return ""
}

func CallRequestFromJSON(data []byte) (*CallRequest, error) {
	c := CallRequest{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CallRequestFromJSONReader(in io.Reader) (*CallRequest, error) {
	c := CallRequest{}
	err := json.NewDecoder(in).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func NewCall(path string) Call {
	c := Call{
		Path: path,
	}
	return c
}

func (c *Call) WithDefault(def Call) Call {
	if c == nil {
		return def
	}
	clone := *c

	if clone.Path == "" {
		clone.Path = def.Path
	}
	if clone.Expand == nil {
		clone.Expand = def.Expand
	}
	if clone.State == nil {
		clone.State = def.State
	}
	return clone
}

func (c *Call) PartialCopy() *Call {
	if c == nil {
		return nil
	}

	clone := *c
	if clone.Expand != nil {
		cloneExpand := *clone.Expand
		clone.Expand = &cloneExpand
	}

	// Only know how to clone map values for State.
	if state, ok := clone.State.(map[string]interface{}); ok {
		cloneState := map[string]interface{}{}
		for k, v := range state {
			cloneState[k] = v
		}
		clone.State = cloneState
	}
	return &clone
}

func (creq *CallRequest) GetValue(name, defaultValue string) string {
	if len(creq.Values) == 0 {
		return defaultValue
	}

	s, ok := creq.Values[name].(string)
	if ok && s != "" {
		return s
	}

	opt, ok := creq.Values[name].(map[string]interface{})
	if ok {
		if v, ok2 := opt["value"].(string); ok2 {
			return v
		}
	}

	return defaultValue
}

func (creq *CallRequest) BoolValue(name string) bool {
	if len(creq.Values) == 0 {
		return false
	}

	isBool := func(v interface{}) (bool, bool) {
		if b, ok := v.(bool); ok {
			return b, true
		}
		if b, ok := creq.Values[name].(string); ok {
			switch b {
			case "true":
				return true, true
			case "false":
				return false, true
			}
		}
		return false, false
	}

	if b, ok := isBool(creq.Values[name]); ok {
		return b
	}
	opt, ok := creq.Values[name].(map[string]interface{})
	if ok {
		if v, ok2 := isBool(opt["value"]); ok2 {
			return v
		}
	}

	return false
}
