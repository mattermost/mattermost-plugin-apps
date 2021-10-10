// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"fmt"
)

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
