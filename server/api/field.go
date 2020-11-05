package api

type FieldType string

const (
	FieldTypeText          = FieldType("text")
	FieldTypeStaticSelect  = FieldType("static_select")
	FieldTypeDynamicSelect = FieldType("dynamic_select")
	FieldTypeBool          = FieldType("bool")
	FieldTypeUser          = FieldType("user")
	FieldTypeChannel       = FieldType("channel")
)

type SelectOption struct {
	Label    string `json:"label"`
	Value    string `json:"value"`
	IconData string `json:"icon_data"`
}

type Field struct {
	// Name is the name of the JSON field to use.
	Name       string    `json:"name"`
	Type       FieldType `json:"type"`
	IsRequired bool      `json:"is_required,omitempty"`

	// Present (default) value of the field
	Value string `json:"value,omitempty"`

	Description string `json:"description,omitempty"`

	// Autocomplete name should be set to either the name of the --flag for
	// named fields, or to $X for positional arguments, where X is 1-based
	// number.
	AutocompleteLabel string `json:"autocomplete_label"`
	AutocompleteHint  string `json:"hint"`
	Position          int    `json:"position"`

	ModalLabel string `json:"modal_label"`

	// Select props
	SelectRefreshOnChangeTo []string       `json:"refresh_on_change_to,omitempty"`
	SelectSourceURL         string         `json:"source_url,omitempty"`
	SelectStaticOptions     []SelectOption `json:"options,omitempty"`

	// Text props
	TextSubtype   string `json:"subtype,omitempty"`
	TextMinLength int    `json:"min_length,omitempty"`
	TextMaxLength int    `json:"max_length,omitempty"`
}
