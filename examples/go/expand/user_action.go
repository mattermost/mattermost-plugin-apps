package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/goapp"
)

var userAction = goapp.NewBindableForm("user-action", apps.Form{
	Title: "Test how Expand works on user actions"
	Header: "TODO",
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

})

actionForm 
