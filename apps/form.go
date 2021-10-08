// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

// Form defines what inputs a Call accepts, and how they can be gathered from
// the user, in Modal and Autocomplete modes.
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
	// Title, Header, and Footer are used for Modals only.
	Title  string `json:"title,omitempty"`
	Header string `json:"header,omitempty"`
	Footer string `json:"footer,omitempty"`

	// A fully-qualified URL, or a path to the form icon.
	// TODO do we default to the App icon?
	Icon string `json:"icon,omitempty"`

	// DeprecatedCall is deprecated in favor of Submit, Source
	DeprecatedCall *Call `json:"call,omitempty"`
	
	// Submit is the call to make when the user clicks a submit button (or enter
	// for a command). A simple call can be specified as a path (string). It
	// will contain no expand/state.
	Submit *Call `json:"submit,omitempty"`

	// Source is the call to make when the form's definition is required (i.e.
	// it has no fields, or needs to be refreshed from the app). A simple call
	// can be specified as a path (string). It will contain no expand/state.
	Source *Call `json:"source,omitempty"`

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
