package apps

type FieldType string
type TextFieldSubtype string

const (
	// Text field. "subtype":"textarea" for multi-line text input (Modal only?).
	// min_length, max_length - self explanatory. Value (default) is expected to be type string, or nil.
	FieldTypeText FieldType = "text"

	// Static select field. The options are specified in "options"
	FieldTypeStaticSelect FieldType = "static_select"

	// Dynamic select's options are fetched by making a call to the Form's Call,
	// in lookup mode. It responds with the list of options to display.
	FieldTypeDynamicSelect FieldType = "dynamic_select"

	// Boolean (checkbox) field.
	FieldTypeBool FieldType = "bool"

	// A mattermost @username.
	FieldTypeUser FieldType = "user"

	// A mattermost channel reference (~name).
	FieldTypeChannel FieldType = "channel"

	// An arbitrary markdown text. Only visible on modals.
	FieldTypeMarkdown FieldType = "markdown"

	TextFieldSubtypeInput     TextFieldSubtype = "input"
	TextFieldSubtypeTextarea  TextFieldSubtype = "textarea"
	TextFieldSubtypeNumber    TextFieldSubtype = "number"
	TextFieldSubtypeEmail     TextFieldSubtype = "email"
	TextFieldSubtypeTelephone TextFieldSubtype = "tel"
	TextFieldSubtypeURL       TextFieldSubtype = "url"
	TextFieldSubtypePassword  TextFieldSubtype = "password"
)

type SelectOption struct {
	// Label is the display name/label for the option's value.
	Label string `json:"label"`

	Value    string `json:"value"`
	IconData string `json:"icon_data"`
}

type Field struct {
	// Name is the name of the JSON field to use.
	Name string `json:"name"`

	Type       FieldType `json:"type"`
	IsRequired bool      `json:"is_required,omitempty"`
	ReadOnly   bool      `json:"readonly,omitempty"`

	// Present (default) value of the field
	Value interface{} `json:"value,omitempty"`

	// Field description. Used in modal and autocomplete.
	Description string `json:"description,omitempty"`

	// Label is used in Autocomplete as the --flag name. It is ignored for
	// positional fields (with AutocompletePosition != 0).
	//
	// TODO: Label should default to Name.
	Label string `json:"label,omitempty"`

	// AutocompleteHint is used in Autocomplete as the hint line.
	AutocompleteHint string `json:"hint,omitempty"`

	// AutocompletePosition means that this is a positional argument, does not
	// have a --flag. If >0, indicates what position this field is in.
	AutocompletePosition int `json:"position,omitempty"`

	// TODO: ModalLabel should default to Label, Name
	ModalLabel string `json:"modal_label"`

	// SelectIsMulti designates whether a select field is a multiselect
	SelectIsMulti bool `json:"multiselect,omitempty"`

	// SelectRefresh means that a change in the value of this select triggers
	// reloading the form. Values of the fields with inputs that are not
	// included in the refreshed form are lost. Not yet implemented for /command
	// autocomplete.
	SelectRefresh bool `json:"refresh,omitempty"`

	// SelectStaticOptions is the list of options to display in a static select
	// field.
	SelectStaticOptions []SelectOption `json:"options,omitempty"`

	// Text props
	TextSubtype   TextFieldSubtype `json:"subtype,omitempty"`
	TextMinLength int              `json:"min_length,omitempty"`
	TextMaxLength int              `json:"max_length,omitempty"`
}
