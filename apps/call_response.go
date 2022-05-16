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

	// Text is used for OK and Error response, and will show the text in the
	// proper output.
	Text string `json:"text,omitempty"`

	// Used in CallResponseTypeOK to return the displayble, and JSON results
	Data interface{} `json:"data,omitempty"`

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
		Text: err.Error(),
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
		Type: CallResponseTypeOK,
		Text: fmt.Sprintf(format, args...),
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

// Error makes CallResponse a valid error, for convenience
func (cresp CallResponse) Error() string {
	if cresp.Type == CallResponseTypeError {
		return cresp.Text
	}
	return ""
}

func (cresp CallResponse) String() string {
	switch cresp.Type {
	case CallResponseTypeError:
		return fmt.Sprint("Error: ", cresp.Text)

	case "", CallResponseTypeOK:
		switch {
		case cresp.Text != "" && cresp.Data == nil:
			return fmt.Sprint("OK: ", cresp.Text)
		case cresp.Text != "" && cresp.Data != nil:
			return fmt.Sprintf("OK: %s; + data type %T", cresp.Text, cresp.Data)
		case cresp.Data != nil:
			return fmt.Sprintf("OK: data type %T, value: %v", cresp.Data, cresp.Data)
		default:
			return "OK: (none)"
		}

	case CallResponseTypeForm:
		return fmt.Sprintf("Form: %v", "omitted for logging")

	case CallResponseTypeCall:
		return fmt.Sprintf("Call: %v", cresp.Call)

	case CallResponseTypeNavigate:
		s := fmt.Sprintf("Navigate to: %q", cresp.NavigateToURL)
		if cresp.UseExternalBrowser {
			s += ", using external browser"
		}
		return s

	default:
		return fmt.Sprintf("?? unknown response type %q", cresp.Type)
	}
}

func (cresp CallResponse) Loggable() []interface{} {
	props := []interface{}{"response_type", string(cresp.Type)}

	switch cresp.Type {
	case CallResponseTypeError:
		props = append(props, "error", cresp.Text)

	case "", CallResponseTypeOK:
		if cresp.Text != "" {
			text := cresp.Text
			if len(text) > 100 {
				text = text[:100] + "...(truncated)"
			}
			props = append(props, "response_text", text)
		}
		if cresp.Data != nil {
			props = append(props, "response_data", "omitted for logging")
		}

	case CallResponseTypeForm:
		if cresp.Form != nil {
			props = append(props, "response_form", "omitted for logging")
		}

	case CallResponseTypeCall:
		if cresp.Call != nil {
			props = append(props, "response_call", cresp.Call.String())
		}

	case CallResponseTypeNavigate:
		props = append(props, "response_url", cresp.NavigateToURL)
		if cresp.UseExternalBrowser {
			props = append(props, "use_external_browser", true)
		}
	}

	return props
}
