// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

import (
	"encoding/json"
)

// Form defines what inputs a Call accepts, and how they can be gathered from
// the user, in Modal and Autocomplete modes.
//
// IMPORTANT: update UnmarshalJSON if this struct changes.
//
// For a Modal, the form defines the modal entirely, and displays it when
// returned in response to a submit Call. Modals are dynamic in the sense that
// they may be refreshed entirely when values of certain fields change, and may
// contain dynamic select fields.
//
// For Autocomplete, a form can be bound to a sub-command. The behavior of
// autocomplete once the subcommand is selected is designed to mirror the
// functionality of the Modal. Some gaps and differences still remain.
//
// A form can be dynamically fetched if it specifies its Source. Source may
// include Expand and State, allowing to create custom-fit forms for the
// context.
type Form struct {
	// Source is the call to make when the form's definition is required (i.e.
	// it has no fields, or needs to be refreshed from the app). A simple call
	// can be specified as a path (string).
	Source *Call `json:"source,omitempty"`

	// Title, Header, and Footer are used for Modals only.
	Title  string `json:"title,omitempty"`
	Header string `json:"header,omitempty"`
	Footer string `json:"footer,omitempty"`

	// A fully-qualified URL, or a path to the form icon.
	// TODO do we default to the App icon?
	Icon string `json:"icon,omitempty"`

	// Submit is the call to make when the user clicks a submit button (or enter
	// for a command). A simple call can be specified as a path (string). It
	// will contain no expand/state.
	Submit *Call `json:"submit,omitempty"`

	// SubmitButtons refers to a field name that must be a FieldTypeStaticSelect
	// or FieldTypeDynamicSelect.
	//
	// In Modal view, the field will be rendered as a list of buttons at the
	// bottom. Clicking one of them submits the Call, providing the button
	// reference as the corresponding Field's value. Leaving this property
	// blank, displays the default "OK".
	//
	// In Autocomplete, it is ignored.
	SubmitButtons string `json:"submit_buttons,omitempty"`

	// Fields is the list of fields in the form.
	Fields []Field `json:"fields,omitempty"`
}

func (f *Form) UnmarshalJSON(data []byte) error {
	stringValue := ""
	err := json.Unmarshal(data, &stringValue)
	if err == nil {
		*f = Form{
			Source: &Call{
				Path: stringValue,
			},
		}
		return nil
	}

	// Need a type that is just like Form, but without UnmarshalJSON
	structValue := struct {
		Source        *Call   `json:"source,omitempty"`
		Title         string  `json:"title,omitempty"`
		Header        string  `json:"header,omitempty"`
		Footer        string  `json:"footer,omitempty"`
		Icon          string  `json:"icon,omitempty"`
		Submit        *Call   `json:"submit,omitempty"`
		SubmitButtons string  `json:"submit_buttons,omitempty"`
		Fields        []Field `json:"fields,omitempty"`
	}{}
	err = json.Unmarshal(data, &structValue)
	if err != nil {
		return err
	}

	*f = Form{
		Source:        structValue.Source,
		Title:         structValue.Title,
		Header:        structValue.Header,
		Footer:        structValue.Footer,
		Icon:          structValue.Icon,
		Submit:        structValue.Submit,
		SubmitButtons: structValue.SubmitButtons,
		Fields:        structValue.Fields,
	}
	return nil
}

func NewFormRef(source *Call) *Form {
	return &Form{Source: source}
}

func NewBlankForm(submit *Call) *Form {
	return &Form{Submit: submit}
}

func (f *Form) IsSubmittable() bool {
	return f != nil && f.Submit != nil
}

func (f *Form) PartialCopy() *Form {
	if f == nil {
		return &Form{}
	}
	clone := *f
	clone.Submit = f.Submit.PartialCopy()
	clone.Source = f.Source.PartialCopy()
	clone.Fields = nil
	for _, field := range f.Fields {
		clone.Fields = append(clone.Fields, *field.PartialCopy())
	}
	return &clone
}
