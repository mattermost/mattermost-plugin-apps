package apps

type Form struct {
	Title  string `json:"title,omitempty"`
	Header string `json:"header,omitempty"`
	Footer string `json:"footer,omitempty"`
	Icon   string `json:"icon,omitempty"`

	Call *Call `json:"call,omitempty"`

	// SubmitButtons refers to a field name that must be a FieldTypeStaticSelect
	// or FieldTypeDynamicSelect.
	//
	// In Modal view, the field will be rendered as a list of buttons at the
	// bottom. Clicking one of them submits the Call, providing the button
	// reference as the corresponding Field's value. Leaving this property
	// blank, displays the default "OK" button.
	//
	// In Autocomplete, it is ignored.
	SubmitButtons string `json:"submit_buttons,omitempty"`

	// Adds a default "Cancel" button in the modal view
	CancelButton  bool `json:"cancel_button,omitempty"`
	SubmitOnCanel bool `json:"submit_on_cancel,omitempty"`

	// DependsOn is the list of field names that when changed force reloading
	// the form. Values of the fields with values that are not included in the
	// refreshed form are lost.
	DependsOn []string `json:"depends_on,omitempty"`

	Fields []*Field `json:"fields,omitempty"`
}
